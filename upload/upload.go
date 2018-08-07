package upload

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/IgaguriMK/sshkeymanager/subcmd"
)

const (
	maxPassRetry = 5
)

func init() {
	subcmd.AddSubCommand(new(Upload))
}

type Upload struct {
	serverName string
}

func (_ *Upload) Cmd() string {
	return "upload"
}

func (_ *Upload) Help() string {
	return "Upload public keys"
}

func (up *Upload) Register(cc *kingpin.CmdClause) {
	cc.Arg("server", "Server name").StringVar(&up.serverName)
}

func (up *Upload) Run() {
	fmt.Println(up.serverName)

	key, err := ReadPrivateKey()
	if err != nil {
		log.Fatal(err)
	}

	config := &ssh.ClientConfig{
		User: "mainuser",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", up.serverName+":52773", config)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		log.Println(err)
	}
	defer session.Close()

	session.Run("cd .ssh && touch authorized_keys && cat authorized_keys | sort | uniq > authorized_keys.uq && chmod 600 authorized_keys.uq && mv authorized_keys.uq authorized_keys")
}

func ReadPrivateKey() (ssh.Signer, error) {
	bs, err := ReadRawPrivateKey()
	if err != nil {
		return nil, err
	}

	key, err := ssh.ParsePrivateKey(bs)
	if err == nil {
		return key, nil
	}

	msg := "Type Password:"

	for i := 0; i < maxPassRetry; i++ {
		pass, err := readPass(msg)
		if err != nil {
			log.Fatal("Failed read password.")
		}

		key, err := ssh.ParsePrivateKeyWithPassphrase(bs, pass)
		if err == nil {
			return key, nil
		}

		msg = "\nFailed, Retype Password:"
	}

	return nil, errors.New("Max retry exceeded")
}

func readPass(msg string) ([]byte, error) {
	fmt.Print(msg)
	// 残念ながらminttyではterminal.ReadPasswordは使えない。
	// 仕方ないので、見えないようにして入力後はパスワードを消すようにしておく。
	// 残念ながら、消される前にコピーするとコピーできてしまう。
	if runtime.GOOS == "windows" && strings.HasPrefix(os.Getenv("TERM"), "xterm") {
		os.Stdout.Write([]byte("\033[8;37;40m"))

		reader := bufio.NewReaderSize(os.Stdin, 1024)
		pass, _, err := reader.ReadLine()

		os.Stdout.Write([]byte("\033[1A"))
		os.Stdout.Write([]byte("\033[2K"))
		os.Stdout.Write([]byte("\033[m"))

		fmt.Println(msg)

		return pass, err
	}

	pass, err := terminal.ReadPassword(int(syscall.Stdin))
	return pass, err
}

func ReadRawPrivateKey() ([]byte, error) {
	home, ok := lookupEnvs("HOME", "USERPROFILE")
	if !ok {
		return nil, errors.New("\"HOME\" is not set")
	}

	bs, err := ioutil.ReadFile(filepath.Join(home, ".ssh/id_rsa"))
	if err != nil {
		return nil, err
	}

	return bs, nil
}

func lookupEnvs(keys ...string) (string, bool) {
	for _, key := range keys {
		v, ok := os.LookupEnv(key)
		if ok {
			return v, true
		}
	}

	return "", false
}
