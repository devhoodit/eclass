package main

import (
	"fmt"

	"github.com/devhoodit/eclass/eclass"
)

func main() {
	e, err := eclass.New("", "")
	if err != nil {
		return
	}
	err = e.AsyncAutoRunLecture()
	fmt.Println(err)
	// e.Pretty_print()
	// subjects, err := e.GetAllSubjects()
	// if err != nil {
	// 	return
	// }

	// for _, subject := range subjects {
	// 	fmt.Printf("%s %s %s", subject.Name, subject.Code, subject.Kj)
	// }

}
