package Search

import "sync"

// import "log"

type Agenda interface {
	Contains(Candidate) bool
	Candidates() Candidates
}

type Problem interface{}
type Candidate interface{}
type Candidates []Candidate

type Interface interface {
	StartItem(p Problem) Candidates
	Clear() Agenda
	Insert(cs chan Candidate, a Agenda) Agenda
	Expand(c Candidate, p Problem) chan Candidate
	Top(a Agenda) Candidate
	GoalTest(p Problem, c Candidate) bool
	TopB(a Agenda, B int) Candidates
}

func Search(b Interface, problem Problem, B int) Candidate {
	candidate, _ := search(b, problem, B, 1, false, false, nil)
	return candidate
}

func SearchConcurrent(b Interface, problem Problem, B int) Candidate {
	candidate, _ := search(b, problem, B, 1, true, false, nil)
	return candidate
}

func SearchConcurrentEarlyUpdate(b Interface, problem Problem, B int, goldSequence []interface{}) (Candidate, Candidate) {
	return search(b, problem, B, 1, true, true, goldSequence)
}

func search(b Interface, problem Problem, B, topK int, concurrent bool, earlyUpdate bool, goldSequence []interface{}) (Candidate, Candidate) {
	var (
		goldValue interface{} = nil
		best      Candidate
		// for early update
		i int
	)

	// candidates <- {STARTITEM(problem)}
	candidates := b.StartItem(problem)
	// loop do
	for {
		// log.Println()
		// log.Println()
		// log.Println("At gold sequence", i)
		if earlyUpdate {
			goldValue = goldSequence[i]
			// log.Println("Gold:", goldValue)
		}
		// agenda <- CLEAR(agenda)
		agenda := b.Clear()
		var wg sync.WaitGroup
		// for each candidate in candidates
		for _, candidate := range candidates {
			wg.Add(1)
			go func(ag Agenda, cand Candidate) {
				defer wg.Done()
				// agenda <- INSERT(EXPAND(candidate,problem),agenda)
				agenda = b.Insert(b.Expand(cand, problem), ag)
			}(agenda, candidate)
			if !concurrent {
				wg.Wait()
			}
		}
		wg.Wait()

		// for each candidate in candidates
		// for _, candidate := range candidates {
		// 	// agenda <- INSERT(EXPAND(candidate,problem),agenda)
		// 	agenda = b.Insert(b.Expand(candidate, problem), agenda)
		// }

		// best <- TOP(AGENDA)
		best = b.Top(agenda)
		// log.Println("Best:", best)
		// log.Println()
		// log.Println("Agenda:")
		for i, _ := range agenda.Candidates() {
			if i == B {
				// log.Println("----- end beam -----")
			}
			// log.Println(c)
		}
		// early update
		if earlyUpdate && !agenda.Contains(goldValue) {
			// log.Println("Early update!")
			return best, goldValue
		}

		// if GOALTEST(problem,best)
		if b.GoalTest(problem, best) {
			// return best
			return best, goldValue
		}
		// candidates <- TOP-B(agenda, B)
		candidates = b.TopB(agenda, B)

		// log.Println()
		// log.Println("Candidates:")
		// for i, _ := range candidates {
		// 	if i == B {
		// log.Println("----- end beam -----")
		// }
		// log.Println(c)
		// }

		// if we're on early update and we've exhausted the gold sequence,
		// break and return a nil gold value
		i++
		if earlyUpdate {
			if i >= len(goldSequence) {
				break
			}
		}
	}
	return best, goldValue
}