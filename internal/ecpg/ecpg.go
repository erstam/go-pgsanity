/* Copyright (c) 2021 Eric St-Amand
See LICENSE for details. */

package ecpg

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
)

const NoneInt = -1

func FromRawSQLFilePath(f string) []byte {
	var err error
	var fh *os.File
	var buf []byte
	if fh, err = os.Open(f); err != nil {
		log.Fatalf("pgsanity: cannot open file %s: %v", f, err)
	}
	defer func(fh *os.File) {
		err := fh.Close()
		if err != nil {
			log.Fatalf("pgsanity: cannot close file %s: %v", f, err)
		}
	}(fh)
	if buf, err = io.ReadAll(fh); err != nil {
		log.Fatalf("pgsanity: cannot read file %s: %v", f, err)
	}

	return prepareSql(buf)
}

type reader interface {
	ReadString(delim byte) (line string, err error)
}

func read(r reader, delim []byte) (line []byte, err error) {
	for {
		s := ""
		s, err = r.ReadString(delim[len(delim)-1])
		if err != nil {
			line = nil
			return
		}

		line = append(line, []byte(s)...)
		if bytes.HasSuffix(line, delim) {
			return line[:len(line)-len(delim)], nil
		}
	}
}

func prepareSql(buf []byte) []byte {

	var result = bytes.NewBuffer(make([]byte, 0, 2*len(buf)))

	inStatement := false
	inLineComment := false
	inBlockComment := false

	for _, segment := range parseSegments(buf) {

		start := segment.Start
		end := segment.End
		contents := segment.Content

		precontents := ""
		startString := start

		if !inStatement && !inLineComment && !inBlockComment {
			// Currently not in any block
			if start != "--" && start != "/*" && len(bytes.TrimSpace(contents)) > 0 {
				inStatement = true
				precontents = "EXEC SQL "
			}
		}

		if start == "/*" {
			inBlockComment = true
		} else if start == "--" && !inBlockComment {
			inLineComment = true
			if !inStatement {
				startString = "//"
			}
		}

		result.Write([]byte(startString))
		result.Write([]byte(precontents))
		result.Write(contents)

		if !inLineComment && !inBlockComment && inStatement && end == ";" {
			inStatement = false
		}

		if inBlockComment && end == "*/" {
			inBlockComment = false
		}

		if inLineComment && end == "\n" {
			inLineComment = false
		}
	}

	return result.Bytes()
}

type Segment struct {
	Start   string
	End     string
	Content []byte
}

func (s Segment) String() string {
	return fmt.Sprintf("Segment: Start bookend: %s, End bookend: %s, content: %s", s.Start, s.End, s.Content)
}

func parseSegments(buf []byte) []Segment {
	segments := make([]Segment, 0, 10)

	bookends := []string{"\n", ";", "--", "/*", "*/"}
	lastBookendFound := ""
	start := 0

	for {
		end, bookend := getNextOccurence(buf[start:], bookends)
		if end == NoneInt {
			// This is probably the last Segment
			segment := Segment{
				Start:   lastBookendFound,
				End:     "",
				Content: buf[start:],
			}
			segments = append(segments, segment)
			start++
		} else {
			end = start + end
			segment := Segment{
				Start:   lastBookendFound,
				End:     bookend,
				Content: buf[start:end],
			}
			segments = append(segments, segment)
			start = end + len(bookend)
			lastBookendFound = bookend
		}

		if start > len(buf) {
			break
		}
	}

	return segments
}

func getNextOccurence(buf []byte, bookends []string) (end int, bookend string) {

	end = len(buf)
	var line []byte
	var err error
	var chosenBookend = ""

	for _, bookend := range bookends {
		if line, err = read(bytes.NewBuffer(buf), []byte(bookend)); err != nil {
			if err != io.EOF {
				log.Fatalf("pgsanity: Internal error while searching for next Segment: %v", err)
			}
		}
		if line != nil {
			if len(line) < end {
				end = len(line)
				chosenBookend = bookend
			}
		}
	}

	if end == len(buf) {
		// No new occurence
		return NoneInt, chosenBookend
	} else {
		return end, chosenBookend
	}
}

func CheckSyntax(content []byte) (err error) {

	//fmt.Println(string(content))

	ecpgCmd := exec.Command("ecpg", "-o", "-", "-")

	var cmdIn io.WriteCloser
	var cmdOut io.ReadCloser
	var cmdErr io.ReadCloser
	var outBytes []byte

	if cmdIn, err = ecpgCmd.StdinPipe(); err != nil {
		return
	}
	if cmdOut, err = ecpgCmd.StdoutPipe(); err != nil {
		return
	}
	if cmdErr, err = ecpgCmd.StderrPipe(); err != nil {
		return
	}
	if err = ecpgCmd.Start(); err != nil {
		return
	}

	// Run the write on stdin in a goroutine to allow the main thread to start reading output pipes.
	// Otherwise, on large SQL files this can cause a deadlock in the io with the subprocess.
	go func() {
		if _, err = cmdIn.Write(content); err != nil {
			return
		}
		if err = cmdIn.Close(); err != nil {
			return
		}
	}()

	if outBytes, err = io.ReadAll(cmdOut); err != nil {
		return
	}
	//if len(outBytes) > 0 {
	//    log.Println(string(outBytes))
	//}

	if outBytes, err = io.ReadAll(cmdErr); err != nil {
		return
	}
	if len(outBytes) > 0 {
		log.Println(string(outBytes))
	}

	if err = ecpgCmd.Wait(); err != nil {
		return
	}

	return nil
}
