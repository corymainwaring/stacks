package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"strings"
	"time"
)

/*
A -> B -> C -> D
Score by number of dependencies

A -> B -> A
Detect circular dependency

Done section*
Dependency only counts if not done

*/

type Stack struct {
	Sections map[string][]*StackEntry
}

func (s *Stack) sort() *Stack {
	return s
}

type StackEntry struct {
	Title        string
	Description  string
	Id           string
	Section      string
	Tags         []string
	Dependencies []*StackEntry
}

func (s *StackEntry) CountDependencies() int {
	sum := 0
	visitedEntries := make([]*StackEntry, 0, len(s.Dependencies))
	for _, v := range s.Dependencies {
		deps, dependencies := v.CountDependenciesWithCircularTracking(visitedEntries)
		sum += deps
		visitedEntries = dependencies
	}
	return len(s.Dependencies) + sum
}

func (s *StackEntry) CountDependenciesWithCircularTracking(visitedEntries []*StackEntry) (int, []*StackEntry) {
	sum := 0
	visitedEntries = append(visitedEntries, s.Dependencies...)
	for _, v := range s.Dependencies {
		found := false
		for _, alreadySeen := range visitedEntries {
			if alreadySeen == v {
				found = true
			}
		}
		if found {
			continue
		}
		deps, dependencies := v.CountDependenciesWithCircularTracking(visitedEntries)
		sum += deps
		visitedEntries = dependencies
	}
	return len(s.Dependencies) + sum, visitedEntries
}

type ByLeastDependent []*StackEntry

func (a ByLeastDependent) Len() int      { return len(a) }
func (a ByLeastDependent) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByLeastDependent) Less(i, j int) bool {
	return a[i].CountDependencies() < a[j].CountDependencies()
}

func NewStack() *Stack {
	result := new(Stack)
	result.Sections = make(map[string][]*StackEntry, 0)
	return result
}

func (s Stack) Format(f fmt.State, c rune) {
	for section, entries := range s.Sections {
		f.Write([]byte("# " + section + "\n\n"))
		for _, entry := range entries {
			f.Write([]byte("## " + entry.Title + " [" + entry.Id + "]\n\n"))
			f.Write([]byte(entry.Description + "\n\n"))
			if entry.Dependencies != nil && len(entry.Dependencies) > 0 {
				f.Write([]byte("Dependent On:"))
				for _, dep := range entry.Dependencies {
					f.Write([]byte(" [" + dep.Id + "]"))
				}
				f.Write([]byte("\n\n"))
			}
		}
	}
}

func LoadStack(f io.Reader) *Stack {
	result := NewStack()

	data, err := io.ReadAll(f)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	idx := 0
	descriptionStart := -1
	str := string(data)
	var curEntry *StackEntry = nil
	var curSection string = ""
	dependencyMap := make(map[string][]string)
	for idx < len(str) {
		if after, found := strings.CutPrefix(str[idx:], "##"); found {
			if curEntry == nil {
				curEntry = new(StackEntry)
				curEntry.Section = curSection
			} else {
				if descriptionStart > -1 {
					curEntry.Description = strings.TrimSpace(str[descriptionStart:idx])
					descriptionStart = -1
				}
				result.Sections[curEntry.Section] = append(result.Sections[curEntry.Section], curEntry)
				curEntry = new(StackEntry)
				curEntry.Section = curSection
			}
			titleLine, rest, _ := strings.Cut(after, "\n")
			title, id, _ := strings.Cut(titleLine, "[")
			id = strings.TrimRight(strings.TrimSpace(id), "]")
			str, idx = rest, -1
			descriptionStart = 0
			curEntry.Title = strings.TrimSpace(title)
			if len(id) == 0 {
				id = generateId(8)
			}
			curEntry.Id = id
		} else if after, found = strings.CutPrefix(str[idx:], "#"); found {
			if descriptionStart > -1 {
				curEntry.Description = strings.TrimSpace(str[descriptionStart:idx])
				descriptionStart = -1
			}
			curSection, after, _ = strings.Cut(after, "\n")
			result.Sections[curSection] = make([]*StackEntry, 0)
			str, idx = after, -1
		} else if after, found = strings.CutPrefix(str[idx:], "Dependent On:"); found {
			if descriptionStart > -1 {
				curEntry.Description = strings.TrimSpace(str[descriptionStart:idx])
				descriptionStart = -1
			}
			dependencyStr, after, _ := strings.Cut(after, "\n")
			dependencyIds := strings.Split(dependencyStr, " ")
			dependencyMap[curEntry.Id] = make([]string, 0, len(dependencyIds))
			for _, id := range dependencyIds {
				dependency := strings.Trim(id, "[] ")
				if len(dependency) > 0 {
					dependencyMap[curEntry.Id] = append(dependencyMap[curEntry.Id], dependency)
				}
			}
			str, idx = after, -1
		}
		idx++
	}

	if curEntry != nil {
		if descriptionStart > -1 {
			curEntry.Description = strings.TrimSpace(str[descriptionStart:idx])
			descriptionStart = -1
		}
		if _, ok := result.Sections[curSection]; !ok {
			result.Sections[curSection] = make([]*StackEntry, 0)
		}
		result.Sections[curSection] = append(result.Sections[curSection], curEntry)
	}

	for _, entries := range result.Sections {
		for _, entry := range entries {
			if depIds, ok := dependencyMap[entry.Id]; ok {
				entry.Dependencies = make([]*StackEntry, 0, len(depIds))
				for _, depId := range depIds {
					for _, entries := range result.Sections {
						for _, depEntry := range entries {
							if depEntry.Id == depId {
								entry.Dependencies = append(entry.Dependencies, depEntry)
								break
							}
						}
					}
				}
			}
		}
	}

	return result
}

// Copied from:
// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
const letterBytes = "abcdefghijkmnopqrstuvwxyz2345678"
const (
	letterIdxBits = 5                    // 5 bits to represent a letter index since letterBytes has 32 entries
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func generateId(n int) string {
	var src = rand.NewSource(time.Now().UnixNano())
	sb := strings.Builder{}
	sb.Grow(n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}
