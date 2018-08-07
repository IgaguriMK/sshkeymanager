package sshkeymanager

import (
	"github.com/IgaguriMK/sshkeymanager/subcmd"

	_ "github.com/IgaguriMK/sshkeymanager/upload"
)

func Main() {
	subcmd.RunApp("sshkeymanager", "Manage SSH keys")
}
