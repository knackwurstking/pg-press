package main

import "errors"

//import j "github.com/knackwurstking/pg-press/scripts/convert_to_json"

func main() {
	// TODO: convert "./images" and "./json" to new SQL database format "..."
	// 	- "sql/tool.sqlite"	    : metal_sheets, tool_regenerations, tools
	//  - "sql/press.sqlite"    : cycles, presses, press_regenerations
	//  - "sql/note.sqlite"	    : notes
	//  - "sql/user.sqlite"	    : cookies, users
	//  - "sql/reports.sqlite"	: trouble_reports

	var err error

	if err = toolData(); err != nil {
		panic("failed to convert tool data: " + err.Error())
	}

	if err = pressData(); err != nil {
		panic("failed to convert press data: " + err.Error())
	}

	if err = noteData(); err != nil {
		panic("failed to convert note data: " + err.Error())
	}

	if err = userData(); err != nil {
		panic("failed to convert user data: " + err.Error())
	}

	if err = reportsData(); err != nil {
		panic("failed to convert reports data: " + err.Error())
	}
}

func toolData() error {
	return errors.New("not implemented yet")
}

func pressData() error {
	return errors.New("not implemented yet")
}

func noteData() error {
	return errors.New("not implemented yet")
}

func userData() error {
	return errors.New("not implemented yet")
}

func reportsData() error {
	return errors.New("not implemented yet")
}
