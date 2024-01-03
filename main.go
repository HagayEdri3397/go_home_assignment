package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

// Form represents the structure of a form.
type Form struct {
	Name      string     `json:"name" yaml:"name"`
	Questions []Question `json:"questions" yaml:"questions"`
}

// Question represents a question in the form.
type Question struct {
	Text     string   `json:"question" yaml:"question"`
	Answer   string   `json:"answer,omitempty" yaml:"answer,omitempty"`
	Options  []string `json:"options,omitempty" yaml:"options,omitempty"`
	Required bool     `json:"required,omitempty" yaml:"required,omitempty"`
}

const (
	ImportForm = "1"
	FillForm   = "2"
)

var selectedForm *Form

func main() {
	fmt.Println("Welcome, choose an action:")
START:
	fmt.Println("1. Import a form")
	fmt.Println("2. Fill in a form")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	action := scanner.Text()

	switch action {
	case ImportForm:
	IMPORT_FILE_LOOP:
		fmt.Println("Enter the path to the form:")
		scanner.Scan()
		formPath := scanner.Text()
		form, err := importForm(formPath)
		if err != nil {
			fmt.Println("Error importing form:", err)
			goto START
		}
		selectedForm = form
		fmt.Println("Form imported.")
		fmt.Println("Choose an action:")
		fmt.Println("1. Import a form")
		fmt.Println("2. Fill in a form")

		scanner.Scan()
		action = scanner.Text()
		switch action {
		case ImportForm:
			goto IMPORT_FILE_LOOP
		case FillForm:
			if err := fillForm(form); err != nil {
				fmt.Println(err.Error())
				goto START
			}
		default:
			fmt.Println("Invalid option. please choose again")
			goto START
		}

	case FillForm:
		if selectedForm == nil {
			fmt.Println("First, a form must be imported")
			goto START
		}
	default:
		fmt.Println("Invalid option. please choose again")
		goto START
	}
}

func importForm(path string) (*Form, error) {
	fileContent, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var form Form

	if strings.HasSuffix(path, ".json") {
		err = json.Unmarshal(fileContent, &form)
	} else if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
		err = yaml.Unmarshal(fileContent, &form)
	} else {
		return nil, fmt.Errorf("unsupported file format")
	}

	if err != nil {
		return nil, err
	}

	return &form, nil
}

func fillForm(form *Form) error {
	for i, question := range form.Questions {
		fmt.Printf("%d. %s\n", i+1, question.Text)

		if len(question.Options) > 0 {
			fmt.Printf("Options: %s\n", strings.Join(question.Options, ", "))
		}

		if question.Required {
			fmt.Println("This question is required.")
		}

	ANSWER:
		fmt.Print("Your answer: ")

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		answer := strings.ToLower(scanner.Text())

		legalAnswer := true
		if question.Options != nil {
			legalAnswer = false

			// Validate the answer based on the question options
			for _, option := range question.Options {
				if option == answer {
					legalAnswer = true
				}
			}
		}

		if legalAnswer {
			form.Questions[i].Answer = answer
		} else {
			fmt.Println("Answer not possible, try again")
			goto ANSWER
		}
	}

	// Print the filled form as JSON
	filledFormJSON, err := json.MarshalIndent(form, "", "  ")
	if err != nil {
		return errors.New(fmt.Sprint("Error formatting filled form:", err))
	}
	fmt.Println("Thank you for filling the form! Here is the filled form:")
	fmt.Println(string(filledFormJSON))
	// Write the filled form to a new file
	submittedFileName := form.Name + "_submitted.json"
	err = writeFormToFile(submittedFileName, form)
	if err != nil {
		fmt.Println("Error writing filled form to file:", err)
	}
	return nil
}

func writeFormToFile(fileName string, form *Form) error {
	formJSON, err := json.MarshalIndent(form, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(fileName, formJSON, 0644)
	if err != nil {
		return err
	}

	fmt.Printf("Filled form written to file: %s\n", fileName)
	return nil
}
