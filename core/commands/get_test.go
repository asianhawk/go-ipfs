package commands

import (
	"context"
	"fmt"
	"testing"

	cmds "gx/ipfs/Qmb7ARH13wPpsyZSXEDeF8aUqSf1ixmWDuD9eNoHUFM5bW/go-ipfs-cmds"
	cmdkit "gx/ipfs/Qmde5VP1qUkyQXKCfmEUA7bP64V2HAptbJ7phuPp7jXWwg/go-ipfs-cmdkit"
)

func TestGetOutputPath(t *testing.T) {
	cases := []struct {
		args    []string
		opts    cmdkit.OptMap
		outPath string
	}{
		{
			args: []string{"/ipns/multiformats.io/"},
			opts: map[string]interface{}{
				"output": "takes-precedence",
			},
			outPath: "takes-precedence",
		},
		{
			args: []string{"/ipns/multiformats.io/", "some-other-arg-to-be-ignored"},
			opts: cmdkit.OptMap{
				"output": "takes-precedence",
			},
			outPath: "takes-precedence",
		},
		{
			args:    []string{"/ipns/multiformats.io/"},
			outPath: "multiformats.io",
			opts:    cmdkit.OptMap{},
		},
		{
			args:    []string{"/ipns/multiformats.io/logo.svg/"},
			outPath: "logo.svg",
			opts:    cmdkit.OptMap{},
		},
		{
			args:    []string{"/ipns/multiformats.io", "some-other-arg-to-be-ignored"},
			outPath: "multiformats.io",
			opts:    cmdkit.OptMap{},
		},
	}

	_, err := GetCmd.GetOptions([]string{})
	if err != nil {
		t.Fatalf("error getting default command options: %v", err)
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%s-%d", t.Name(), i), func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			req, err := cmds.NewRequest(ctx, []string{}, tc.opts, tc.args, nil, GetCmd)
			if err != nil {
				t.Fatalf("error creating a command request: %v", err)
			}

			if outPath := getOutPath(req); outPath != tc.outPath {
				t.Errorf("expected outPath %s to be %s", outPath, tc.outPath)
			}
		})
	}
}
