package parse

import (
	"fmt"
	"io"
	"log"
	"strings"

	"golang.org/x/net/html"
)

// TODO:
// 1. break the long if-else logic in switch-case as much as you can
// 2. parse the semester count from html too (not logically), although difficult
//    to parse it'll make easier to keep track of back sems (example id:
//    "ctl05_ctl00_lblSemesterId")
// 3. add attribute in Semester to mark a sem as back sem, this will allow for
//    caller function to sort out back semesters easily
// 4. If necessary add more attributes in Semester to total subjects, DOD etc.
// 5. If necessary add logic to parse out student details too
// 6. Try to break that gigantic loop into small functions, try to see if
//    sharing the tokenizer between multiple functions is good idea or not
// 7. Maybe closure or functional options could come in handy for task 5. Look
//    for those
// 8. While breaking the logic try to see what could be formed as public
//    methods for Student
// 9. Try out "goQuery", and see if it'll be good to use it instead of tokenizer

type Subject struct {
	Code string
	Name string
	Type string
}

type Mark struct {
	Subject
	Internal string
	External string
	BackPaper string
	Grade string
}

type Semester struct {
	Count uint
	Marks []Mark
}

type Student struct {
	Sems []Semester
}

var idFormatPrefix string
var idFormatSuffix string
var backSessionPrefix string
var backSessionSuffix string
var yearCounter int
var semCounter int
var backCounter int
var isBack bool

func Parse(r io.Reader) {
	yearCounter = 4
	semCounter = 0
	backCounter = 0
	idFormatSuffix = "_ctl00_grdViewSubjectMarksheet"
	idFormatPrefix = "ctl0%d_ctl0%d%s"
	backSessionPrefix = "ctl0%d%s"
	backSessionSuffix = "_lblSession"
	idFormat := fmt.Sprintf(idFormatPrefix, yearCounter, semCounter % 2, idFormatSuffix)
	backSessionFormat := fmt.Sprintf(backSessionPrefix, yearCounter, backSessionSuffix)
	_ = backSessionFormat

	// try parse logic

	/* hdoc, err := html.Parse(r)
	if err != nil {
		log.Fatalf("parsing error: %v", err)
	}
	_ = hdoc */

	student := Student{
		Sems: make([]Semester, 0, 10),
	}

	var sem Semester
	var mark Mark
	colCounter := 0
	innerData := []string{}

	desiredTable := false
	tokenizer := html.NewTokenizer(r)
	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			err := tokenizer.Err()
			if err == io.EOF {
				// put some warning here
				break
			}
			// otherwise, the html is malformed
			log.Fatalf("html is malformed: %v", err)
		}
		if tokenType == html.StartTagToken {
			token := tokenizer.Token()
			if token.Data == "table" {
				for _, a := range token.Attr {
					// if matches process the table
					if a.Key == "id" && a.Val == idFormat {
						// fmt.Println("before:", idFormat)
						desiredTable = true
						semCounter++
						if semCounter % 2 == 0 {
							yearCounter++
							backSessionFormat = fmt.Sprintf(backSessionPrefix, yearCounter, backSessionSuffix)
						}
						idFormat = fmt.Sprintf(idFormatPrefix, yearCounter, semCounter % 2, idFormatSuffix)
						// fmt.Println("after:", idFormat)
						break
					}
				}
			} else if token.Data == "span" {
				for _, a := range token.Attr {
					// fmt.Println("backSessionFormat:", backSessionFormat)
					if a.Key == "id" && a.Val == backSessionFormat {
						fmt.Println("found")
						tokenizer.Next() // skip StartTagToken <b> tag
						tokenizer.Next() // skip TextToken inside </b> tag
						tokenizer.Next() // skip EndTagToken </b>
						inner := tokenizer.Next()
						if inner == html.TextToken {
							text := strings.TrimSpace(string(tokenizer.Text()))
							if strings.Contains(text, "BACK") {
								isBack = true
								backCounter += 2
							} else {
								isBack = false
							}
						}
						break
					}
				}
			}
			if (desiredTable) {
				if token.Data == "tbody" {
					// fmt.Println("new sem, len:", len(student.Sems))
					if len(student.Sems) < semCounter {
						sem = Semester{
							Count: uint(semCounter) - uint(backCounter),
							Marks: make([]Mark, 0, 10),
						}
						if isBack {
							if semCounter % 2 == 1 {
								sem.Count = uint(semCounter) - uint(backCounter) + 1
							}
						}
					}
				}
				if token.Data == "tr" {
					colCounter = 0
					mark = Mark{}
					innerData = []string{}
				}
				if token.Data == "td" {
					inner := tokenizer.Next()
					if inner == html.TextToken {
						text := string(tokenizer.Text())
						inner = tokenizer.Next()
						if tokenizer.Token().Data == "span" {
							inner = tokenizer.Next()
							if inner == html.TextToken {
								text = string(tokenizer.Text())
							}
						}
						innerData = append(innerData, strings.TrimSpace(text))
						colCounter++
					}
					if colCounter == 7 {
						mark.Code = innerData[0]
						mark.Name = innerData[1]
						mark.Type = innerData[2]
						mark.Internal = innerData[3]
						mark.External = innerData[4]
						mark.BackPaper = innerData[5]
						mark.Grade = innerData[6]
						sem.Marks = append(sem.Marks, mark)
					}
				}
			}
		}
		if tokenType == html.EndTagToken {
			token := tokenizer.Token()
			if token.Data == "table" && desiredTable {
				student.Sems = append(student.Sems, sem)
				desiredTable = false
			}
		}
	}

	fmt.Println("len sem:", len(student.Sems))
	for _, sem := range student.Sems {
		fmt.Println("\n==> Semester count: ", sem.Count)
		for _, mark := range sem.Marks {
			fmt.Printf("\n\tName: %s\n\tCode: %s\n\tType: %s\n", mark.Name, mark.Code, mark.Type)
			fmt.Printf("\t\tInt: %s, Ext: %s, Back: %s, Grade: %s\n", mark.Internal, mark.External, mark.BackPaper, mark.Grade)
		}
		fmt.Println("----------------------------------------------------------")
	}
}

// tableNodes returns the <table> Node which contain the result data
func tableNodes(n *html.Node) []*html.Node {
	if n.Type == html.ElementNode && n.Data == "table" {
		idFormat := fmt.Sprintf(idFormatPrefix, yearCounter, semCounter, idFormatSuffix)
		for _, a := range n.Attr {
			if a.Key == "id" && a.Val == idFormat {
				if semCounter == 0 {
					semCounter = 1
				} else {
					yearCounter++
					semCounter = 0
				}
				return []*html.Node{n}
			}
		}
	}

	var nodes []*html.Node

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		nodes = append(nodes, tableNodes(c)...)
	}

	return nodes
}



/* // using tokenizer to parse the tree
tokenizer := html.NewTokenizer(r)

// loop within tokenizer until result or error
for {
	tokenType := tokenizer.Next()

	// check for error token
	if tokenType == html.ErrorToken {
		err := tokenizer.Err()
		if err == io.EOF {
			// put some warning here
			break
		}
		// otherwise, the html is malformed
		log.Fatalf("html is malformed: %v", err)
	}

	if tokenType == html.StartTagToken {
		// get the token
		token := tokenizer.Token()
		// check the name of token
		if token.Data == "table" {
			fmt.Println(": found a table")
			for _, a := range token.Attr {
				if a.Key == "id" && a.Val == idFormat {
					fmt.Printf("=> Found the needed table: %d\n", foundCount)
					foundCount += 1
					if semCounter == 0 {
						semCounter = 1
					} else {
						semCounter = 0
						yearCounter++
					}
					idFormat = fmt.Sprintf(idFormatPrefix, yearCounter, semCounter, idFormatSuffix)
				}
			}
		}
	}
} */


// parsing approach
/* var f func(n *html.Node)

f = func(n *html.Node) {
	if n.Type == html.ElementNode && n.Data == "table" {
		fmt.Println(": found a table")
		for _, a := range n.Attr {
			if a.Key == "id" && a.Val == idFormat {
				fmt.Printf("=> Found the needed table: %d\n", foundCount)
				foundCount += 1
				if semCounter == 0 {
					semCounter = 1
				} else {
					semCounter = 0
					yearCounter++
				}
				idFormat = fmt.Sprintf(idFormatPrefix, yearCounter, semCounter, idFormatSuffix)
				break
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		f(c)
	}
}

f(hdoc) */
