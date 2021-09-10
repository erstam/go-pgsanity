/* Copyright (c) 2021 Eric St-Amand
See LICENSE for details. */

package pgsanity

import (
	"github.com/erstam/go-pgsanity/internal/args"
	"github.com/erstam/go-pgsanity/internal/ecpg"
	"log"
	"os"
	"path/filepath"
)

func Run() {
	input := args.Parse()
	input, _ = filepath.Abs(input)
	fileInfo, err := os.Stat(input)
	if err != nil {
		log.Fatalf("pgsanity: file not found: %v", err)
	}
	if fileInfo.IsDir() {
		checkDir(input)
	} else {
		checkFile(input)
	}
}

func isSqlFile(f string) bool {
	return filepath.Ext(f) == ".sql"
}

func ensureSql(f string) {
	if !isSqlFile(f) {
		log.Fatalf("file %s does not have .sql extension in its filename.", f)
	}
}

func checkDir(dir string) {
	err := filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || !isSqlFile(path) {
				return nil
			}
			checkFile(path)
			return nil
		})
	if err != nil {
		log.Fatalf("pgsanity: error while checking file: %v", err)
	}
}

func checkFile(sqlFile string) {
	ensureSql(sqlFile)
	log.Printf("checking %s", sqlFile)
	err := ecpg.CheckSyntax(ecpg.FromRawSQLFilePath(sqlFile))
	if err != nil {
		log.Fatalf("pgsanity: error checking syntax of file %s: %v", sqlFile, err)
	}
}
