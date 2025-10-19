package executor

import (
	"testing"
)

func TestValidateCommand(t *testing.T) {
	tests := []struct {
		name    string
		command string
		wantErr bool
	}{
		{"valid command", "echo hello", false},
		{"command with flags", "ls -la", false},
		{"complex command", "docker run -it --rm ubuntu bash", false},
		{"command with pipes", "cat file.txt | grep test", false},
		{"command with redirection", "echo test > output.txt", false},
		{"multiline command", "echo line1 && echo line2", false},
		{"empty command", "", true},
		{"whitespace only", "   ", true},
		{"tab only", "\t", true},
		{"newline only", "\n", true},
		{"mixed whitespace", " \t\n ", true},
		{"carriage return and newline", "\r\n", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCommand(tt.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.wantErr {
				if !contains(err.Error(), "command cannot be") {
					t.Errorf("validateCommand() error message = %q, expected to contain 'command cannot be'", err.Error())
				}
			}
		})
	}
}
