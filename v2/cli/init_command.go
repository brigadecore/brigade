package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/brigadecore/brigade-foundations/file"
	"github.com/brigadecore/brigade/sdk/v3"
	"github.com/urfave/cli/v2"
)

var initCommand = &cli.Command{
	Name:  "init",
	Usage: "Initialize a new Brigade project",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    flagID,
			Aliases: []string{"i"},
			Usage: "Name of the Brigade project to initialize " +
				"(required)",
			Required: true,
		},
		&cli.StringFlag{
			Name:    flagLanguage,
			Aliases: []string{"l"},
			Usage: "Specify the scripting language to use for handling events" +
				" -- JavaScript (js) or TypeScript (ts)",
			Value: "ts",
		},
		&cli.StringFlag{
			Name:    flagGit,
			Aliases: []string{"g"},
			Usage: "Optionally specify a git repository where the" +
				" event handling script can be found",
		},
	},
	Action: initialize,
}

// nolint: gocyclo
func initialize(c *cli.Context) error {

	// Initialize fields to fill in the template yaml files
	fields := struct {
		ProjectID   string
		Language    string
		GitCloneURL string
		Script      string
	}{
		ProjectID:   c.String(flagID),
		Language:    strings.ToLower(c.String(flagLanguage)),
		GitCloneURL: c.String(flagGit),
		Script:      "",
	}

	err := sdk.ValidateProjectID(fields.ProjectID)
	if err != nil {
		return err
	}

	if fields.GitCloneURL != "" {
		if err = sdk.ValidateGitCloneURL(fields.GitCloneURL); err != nil {
			return err
		}
	}

	// Check if input language is valid
	fields.Language, err = fileExtensionForLanguage(fields.Language)
	if err != nil {
		return err
	}

	// Check if the .brigade directory exists
	var exists bool
	exists, err = file.EnsureDirectory(".brigade")
	if err != nil {
		return err
	}
	if !exists {
		fmt.Println("Creating .brigade directory...")
	}

	// Generate script based on given language
	var scriptBytes []byte
	switch fields.Language {
	case "ts":
		scriptBytes, err = execTemplate(typeScriptTemplate, fields)
	case "js":
		scriptBytes, err = execTemplate(javaScriptTemplate, fields)
	}
	if err != nil {
		return err
	}
	fields.Script = string(scriptBytes)

	brigadeScriptFileName := fmt.Sprintf("brigade.%s", fields.Language)

	// If git repo is provided, create a separate js/ts file for script
	var overwrite bool
	var fileWrite bool
	if fields.GitCloneURL != "" {
		scriptPath := path.Join(".brigade", brigadeScriptFileName)
		if fileWrite, overwrite, err =
			checkFileOverwrite(scriptPath); err != nil {
			return err
		}
		if fileWrite {
			if err = ioutil.WriteFile( // nolint: gosec
				scriptPath,
				scriptBytes,
				0644,
			); err != nil {
				return err
			}
			if overwrite {
				fmt.Printf("Overwriting %s...\n", scriptPath)
			} else {
				fmt.Printf("Creating %s...\n", scriptPath)
			}
		}

		// Also generate a minimal package.json file
		var packageBytes []byte
		packageBytes, err = execTemplate(packageTemplate, fields)
		if err != nil {
			return err
		}

		packagePath := path.Join(".brigade", "package.json")
		if fileWrite, overwrite, err =
			checkFileOverwrite(packagePath); err != nil {
			return err
		}
		if fileWrite {
			if err = ioutil.WriteFile( // nolint: gosec
				packagePath,
				packageBytes,
				0644,
			); err != nil {
				return err
			}
			if overwrite {
				fmt.Printf("Overwriting %s...\n", packagePath)
			} else {
				fmt.Printf("Creating %s...\n", packagePath)
			}
		}
	}

	// Generate the project.yaml file
	var projectBytes []byte
	projectBytes, err = execTemplate(projectTemplate, fields)
	if err != nil {
		return err
	}

	// Ensure the .brigade/project.yaml path exists.
	// Confirm overwrite if so.
	projectPath := path.Join(".brigade", "project.yaml")
	if fileWrite, overwrite, err = checkFileOverwrite(projectPath); err != nil {
		return err
	}
	if fileWrite {
		if err = ioutil.WriteFile( // nolint: gosec
			projectPath,
			projectBytes,
			0644,
		); err != nil {
			return err
		}
		if overwrite {
			fmt.Printf("Overwriting %s...\n", projectPath)
		} else {
			fmt.Printf("Creating %s...\n", projectPath)
		}
	}

	var notesBytes []byte
	notesBytes, err = execTemplate(notesTemplate, fields)
	if err != nil {
		return err
	}

	// Create a NOTES.txt file
	notesPath := path.Join(".brigade", "NOTES.txt")
	if fileWrite, overwrite, err = checkFileOverwrite(notesPath); err != nil {
		return err
	}
	if fileWrite {
		if err = ioutil.WriteFile( // nolint: gosec
			notesPath,
			notesBytes,
			0644,
		); err != nil {
			return err
		}
		if overwrite {
			fmt.Printf("Overwriting %s...\n", notesPath)
		} else {
			fmt.Printf("Creating %s...\n", notesPath)
		}
	}

	var secretsBytes []byte
	secretsBytes, err = execTemplate(secretsTemplate, fields)
	if err != nil {
		return err
	}

	// Create a secrets.yaml file
	secretsPath := path.Join(".brigade", "secrets.yaml")
	if fileWrite, overwrite, err = checkFileOverwrite(secretsPath); err != nil {
		return err
	}
	if fileWrite {
		if err = ioutil.WriteFile(
			secretsPath,
			secretsBytes,
			0600,
		); err != nil {
			return err
		}
		if overwrite {
			fmt.Printf("Overwriting %s...\n", secretsPath)
		} else {
			fmt.Printf("Creating %s...\n", secretsPath)
		}
	}

	// Add appropriate files to .gitignore
	if err = addToGitIgnore(
		secretsPath,
		".brigade/node_modules/\n",
	); err != nil {
		return err
	}
	fmt.Printf("Adding %s to .gitignore...\n", secretsPath)
	fmt.Printf("Adding .brigade/node_modules/ to .gitignore...\n")

	fmt.Printf("\nPlease refer to %s for next steps.\n", notesPath)

	return nil
}

// addToGitIgnore invokes addLinesToFile, and adds the given paths to
// the root directory's .gitignore
func addToGitIgnore(pathsToIgnore ...string) error {
	fileExists, err := file.Exists(".gitignore")
	if !fileExists {
		fmt.Printf("Creating .gitignore...\n")
	}
	if err != nil {
		return err
	}
	return addLinesToFile("./.gitignore", pathsToIgnore...)
}

// addLinesToFile detects if a given filepath exists. If it does, add
// the given paths to it. If not, create it and add the specified paths
// to it unconditionally
func addLinesToFile(editFilePath string, pathsToAdd ...string) error {
	file, err :=
		os.OpenFile(editFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	for _, path := range pathsToAdd {
		if _, err =
			file.WriteString(fmt.Sprintf("\n%s", path)); err != nil {
			return err
		}
	}
	return nil
}

// fileExtensionForLanguage checks if a given lagnuage is allowed. If it is,
// shorten it to its file extension. Otherwise, return an error.
func fileExtensionForLanguage(language string) (string, error) {
	language = strings.ToLower(language)
	switch language {
	case "ts", "typescript":
		return "ts", nil
	case "js", "javascript":
		return "js", nil
	default:
		return "", fmt.Errorf("unrecognized value %q for --language flag"+
			"(ts or js expected)", language)
	}
}

// checkFileOverwrite checks if the given file exists, and returns two booleans:
// the first indicates whether or not a file will be written, and the second
// indicates whether that file will be created from scratch or will be
// overwritten
func checkFileOverwrite(
	filepath string,
) (bool, bool, error) {
	if fileExists, err := file.Exists(filepath); err != nil {
		return false, false, err
	} else if fileExists {
		response := false
		prompt := &survey.Confirm{
			Message: "A " + path.Base(filepath) + " file already exists in" +
				" your .brigade directory, would you like to overwrite it?",
			Default: false,
		}

		if err = survey.AskOne(prompt, &response); err != nil {
			return false, false, err
		}

		return response, true, nil
	}
	return true, false, nil
}
