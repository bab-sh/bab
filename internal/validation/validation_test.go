package validation

import (
	"strings"
	"testing"
)

func TestValidateString(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		fieldName string
		wantErr   bool
		errMsg    string
	}{
		{"valid string", "hello", "field", false, ""},
		{"string with spaces", "hello world", "field", false, ""},
		{"empty string", "", "field", true, "field cannot be empty"},
		{"whitespace only", "   ", "field", true, "field cannot be only whitespace"},
		{"tab only", "\t", "field", true, "field cannot be only whitespace"},
		{"newline only", "\n", "field", true, "field cannot be only whitespace"},
		{"mixed whitespace", " \t\n ", "field", true, "field cannot be only whitespace"},
		{"custom field name", "", "path", true, "path cannot be empty"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateString(tt.value, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateString() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.wantErr && tt.errMsg != "" {
				if err.Error() != tt.errMsg {
					t.Errorf("ValidateString() error = %q, want %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

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
			err := ValidateCommand(tt.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.wantErr {
				if !strings.Contains(err.Error(), "command cannot be") {
					t.Errorf("ValidateCommand() error message = %q, expected to contain 'command cannot be'", err.Error())
				}
			}
		})
	}
}

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid path", "/home/user/file.txt", false},
		{"relative path", "./config.yaml", false},
		{"simple filename", "Babfile.yaml", false},
		{"path with spaces", "/home/user/my file.txt", false},
		{"empty path", "", true},
		{"whitespace only", "   ", true},
		{"tab only", "\t", true},
		{"newline only", "\n", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.wantErr {
				if !strings.Contains(err.Error(), "path cannot be") {
					t.Errorf("ValidatePath() error message = %q, expected to contain 'path cannot be'", err.Error())
				}
			}
		})
	}
}

func TestValidateNonEmptySlice(t *testing.T) {
	t.Run("non-empty string slice", func(t *testing.T) {
		err := ValidateNonEmptySlice([]string{"a", "b"}, "commands")
		if err != nil {
			t.Errorf("ValidateNonEmptySlice() unexpected error = %v", err)
		}
	})

	t.Run("single element slice", func(t *testing.T) {
		err := ValidateNonEmptySlice([]string{"a"}, "commands")
		if err != nil {
			t.Errorf("ValidateNonEmptySlice() unexpected error = %v", err)
		}
	})

	t.Run("empty string slice", func(t *testing.T) {
		err := ValidateNonEmptySlice([]string{}, "commands")
		if err == nil {
			t.Error("ValidateNonEmptySlice() expected error, got nil")
		}
		if err != nil && err.Error() != "commands cannot be empty" {
			t.Errorf("ValidateNonEmptySlice() error = %q, want %q", err.Error(), "commands cannot be empty")
		}
	})

	t.Run("non-empty int slice", func(t *testing.T) {
		err := ValidateNonEmptySlice([]int{1, 2, 3}, "numbers")
		if err != nil {
			t.Errorf("ValidateNonEmptySlice() unexpected error = %v", err)
		}
	})

	t.Run("empty int slice", func(t *testing.T) {
		err := ValidateNonEmptySlice([]int{}, "numbers")
		if err == nil {
			t.Error("ValidateNonEmptySlice() expected error, got nil")
		}
	})

	t.Run("non-empty interface slice", func(t *testing.T) {
		err := ValidateNonEmptySlice([]interface{}{"a", 1}, "items")
		if err != nil {
			t.Errorf("ValidateNonEmptySlice() unexpected error = %v", err)
		}
	})

	t.Run("empty interface slice", func(t *testing.T) {
		err := ValidateNonEmptySlice([]interface{}{}, "items")
		if err == nil {
			t.Error("ValidateNonEmptySlice() expected error, got nil")
		}
	})
}

func TestValidateDependencyName(t *testing.T) {
	tests := []struct {
		name     string
		dep      string
		index    int
		taskName string
		wantErr  bool
		errMsg   string
	}{
		{"valid dependency", "build", 0, "test", false, ""},
		{"valid with namespace", "lib:compile", 1, "app:build", false, ""},
		{"empty dependency", "", 0, "test", true, `task "test" has empty dependency at index 0`},
		{"empty at index 2", "", 2, "deploy", true, `task "deploy" has empty dependency at index 2`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDependencyName(tt.dep, tt.index, tt.taskName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDependencyName() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.wantErr && tt.errMsg != "" {
				if err.Error() != tt.errMsg {
					t.Errorf("ValidateDependencyName() error = %q, want %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}
