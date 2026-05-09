package brain

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

// Genome represents a single payload chromosome
type Genome struct {
	DNA          []byte   `json:"dna"`
	Fitness      float64  `json:"fitness"`
	Generation   int      `json:"generation"`
	ParentIDs    []string `json:"parent_ids"`
	MutationType string   `json:"mutation_type"`
}

// Population manages genetic evolution
type Population struct {
	TargetType    string    `json:"target_type"`
	Generations   int       `json:"generations"`
	Size          int       `json:"size"`
	EliteRate     float64   `json:"elite_rate"`
	MutationRate  float64   `json:"mutation_rate"`
	CrossoverRate float64   `json:"crossover_rate"`
	Genomes       []*Genome `json:"genomes"`
	BestFitness   float64   `json:"best_fitness"`
}

// NewPopulation creates a genetic population for payload evolution
func NewPopulation(targetType string, seedPayloads []string, size int) *Population {
	if size < 10 {
		size = 50
	}
	pop := &Population{
		TargetType:    targetType,
		Size:          size,
		EliteRate:     0.1,
		MutationRate:  0.3,
		CrossoverRate: 0.7,
		Genomes:       make([]*Genome, 0, size),
	}

	// Initialize with seed payloads
	for _, seed := range seedPayloads {
		pop.Genomes = append(pop.Genomes, &Genome{
			DNA:        []byte(seed),
			Fitness:    0,
			Generation: 0,
		})
	}

	// Fill rest with random mutations of seeds
	for len(pop.Genomes) < size && len(seedPayloads) > 0 {
		seed := seedPayloads[randInt(len(seedPayloads))]
		mutated := pop.mutatePayload(seed)
		pop.Genomes = append(pop.Genomes, &Genome{
			DNA:          []byte(mutated),
			Fitness:      0,
			Generation:   0,
			MutationType: "random_init",
		})
	}

	return pop
}

// Evolve runs one generation of genetic evolution
func (p *Population) Evolve(fitnessFn func(string) float64) {
	// Evaluate fitness
	for _, g := range p.Genomes {
		if g.Fitness == 0 {
			g.Fitness = fitnessFn(string(g.DNA))
		}
	}

	// Sort by fitness (descending)
	for i := 0; i < len(p.Genomes); i++ {
		for j := i + 1; j < len(p.Genomes); j++ {
			if p.Genomes[j].Fitness > p.Genomes[i].Fitness {
				p.Genomes[i], p.Genomes[j] = p.Genomes[j], p.Genomes[i]
			}
		}
	}

	// Track best
	if len(p.Genomes) > 0 && p.Genomes[0].Fitness > p.BestFitness {
		p.BestFitness = p.Genomes[0].Fitness
	}

	// Elitism - keep top performers
	eliteCount := int(float64(p.Size) * p.EliteRate)
	if eliteCount < 2 {
		eliteCount = 2
	}
	newGenomes := make([]*Genome, 0, p.Size)
	for i := 0; i < eliteCount && i < len(p.Genomes); i++ {
		newGenomes = append(newGenomes, &Genome{
			DNA:        append([]byte{}, p.Genomes[i].DNA...),
			Fitness:    p.Genomes[i].Fitness,
			Generation: p.Genomes[i].Generation + 1,
			ParentIDs:  []string{p.Genomes[i].hashID()},
		})
	}

	// Crossover and mutation to fill rest
	for len(newGenomes) < p.Size {
		parent1 := p.tournamentSelect()
		parent2 := p.tournamentSelect()

		var childDNA []byte
		if randFloat() < p.CrossoverRate {
			childDNA = p.crossover(parent1.DNA, parent2.DNA)
		} else {
			childDNA = append([]byte{}, parent1.DNA...)
		}

		if randFloat() < p.MutationRate {
			childDNA = p.mutateDNA(childDNA)
		}

		newGenomes = append(newGenomes, &Genome{
			DNA:        childDNA,
			Fitness:    0,
			Generation: p.Generations + 1,
			ParentIDs:  []string{parent1.hashID(), parent2.hashID()},
		})
	}

	p.Genomes = newGenomes
	p.Generations++
}

// GetBest returns the highest fitness genome
func (p *Population) GetBest() *Genome {
	if len(p.Genomes) == 0 {
		return nil
	}
	best := p.Genomes[0]
	for _, g := range p.Genomes[1:] {
		if g.Fitness > best.Fitness {
			best = g
		}
	}
	return best
}

// GetTop returns top N genomes as strings
func (p *Population) GetTop(n int) []string {
	if n > len(p.Genomes) {
		n = len(p.Genomes)
	}
	result := make([]string, n)
	for i := 0; i < n; i++ {
		result[i] = string(p.Genomes[i].DNA)
	}
	return result
}

// tournamentSelect picks a parent using tournament selection
func (p *Population) tournamentSelect() *Genome {
	tournamentSize := 3
	if tournamentSize > len(p.Genomes) {
		tournamentSize = len(p.Genomes)
	}
	best := p.Genomes[randInt(len(p.Genomes))]
	for i := 1; i < tournamentSize; i++ {
		contender := p.Genomes[randInt(len(p.Genomes))]
		if contender.Fitness > best.Fitness {
			best = contender
		}
	}
	return best
}

// crossover performs single-point crossover between two DNA sequences
func (p *Population) crossover(dna1, dna2 []byte) []byte {
	if len(dna1) == 0 || len(dna2) == 0 {
		return append([]byte{}, dna1...)
	}
	point := randInt(min(len(dna1), len(dna2)))
	child := make([]byte, 0, len(dna1)+len(dna2))
	child = append(child, dna1[:point]...)
	child = append(child, dna2[point:]...)
	return child
}

// mutateDNA applies real mutations to payload DNA
func (p *Population) mutateDNA(dna []byte) []byte {
	mutations := []func([]byte) []byte{
		mutateEncoding,
		mutateCase,
		mutateComments,
		mutateWhitespace,
		mutateSwapTechnique,
		mutateConcat,
		mutateUnicode,
		mutateNullByte,
	}

	mutation := mutations[randInt(len(mutations))]
	return mutation(dna)
}

// mutatePayload applies a single mutation to a string payload
func (p *Population) mutatePayload(payload string) string {
	return string(p.mutateDNA([]byte(payload)))
}

// Real mutation functions
func mutateEncoding(dna []byte) []byte {
	result := make([]byte, 0, len(dna)*3)
	for _, b := range dna {
		switch b {
		case '\'':
			result = append(result, []byte("%27")...)
		case '"':
			result = append(result, []byte("%22")...)
		case ' ':
			result = append(result, []byte("%20")...)
		case '<':
			result = append(result, []byte("%3c")...)
		case '>':
			result = append(result, []byte("%3e")...)
		default:
			result = append(result, b)
		}
	}
	return result
}

func mutateCase(dna []byte) []byte {
	result := make([]byte, len(dna))
	for i, b := range dna {
		if b >= 'a' && b <= 'z' && randFloat() < 0.3 {
			result[i] = b - ('a' - 'A')
		} else if b >= 'A' && b <= 'Z' && randFloat() < 0.3 {
			result[i] = b + ('a' - 'A')
		} else {
			result[i] = b
		}
	}
	return result
}

func mutateComments(dna []byte) []byte {
	comment := "/**/"
	idx := randInt(len(dna))
	result := make([]byte, 0, len(dna)+len(comment))
	result = append(result, dna[:idx]...)
	result = append(result, []byte(comment)...)
	result = append(result, dna[idx:]...)
	return result
}

func mutateWhitespace(dna []byte) []byte {
	result := make([]byte, 0, len(dna)*2)
	for _, b := range dna {
		if b == ' ' && randFloat() < 0.5 {
			result = append(result, []byte("/**/")...)
		} else {
			result = append(result, b)
		}
	}
	return result
}

func mutateSwapTechnique(dna []byte) []byte {
	s := string(dna)
	s = strings.Replace(s, "SELECT", "SEL/**/ECT", -1)
	s = strings.Replace(s, "UNION", "UNI/**/ON", -1)
	s = strings.Replace(s, "AND", "A/**/ND", -1)
	s = strings.Replace(s, "OR", "O/**/R", -1)
	s = strings.Replace(s, "script", "scr\\x69pt", -1)
	return []byte(s)
}

func mutateConcat(dna []byte) []byte {
	s := string(dna)
	replacements := map[string]string{
		"SELECT": "SELE'+'CT",
		"script": "scr" + string(rune(0x00)) + "ipt",
	}
	for old, new := range replacements {
		s = strings.Replace(s, old, new, -1)
	}
	return []byte(s)
}

func mutateUnicode(dna []byte) []byte {
	result := make([]byte, 0, len(dna)*2)
	for _, b := range dna {
		switch b {
		case '<':
			result = append(result, []byte("\\u003c")...)
		case '>':
			result = append(result, []byte("\\u003e")...)
		case '"':
			result = append(result, []byte("\\u0022")...)
		case '\'':
			result = append(result, []byte("\\u0027")...)
		default:
			result = append(result, b)
		}
	}
	return result
}

func mutateNullByte(dna []byte) []byte {
	idx := randInt(len(dna))
	result := make([]byte, 0, len(dna)+1)
	result = append(result, dna[:idx]...)
	result = append(result, 0x00)
	result = append(result, dna[idx:]...)
	return result
}

func (g *Genome) hashID() string {
	return fmt.Sprintf("%x", g.DNA[:min(8, len(g.DNA))])
}

func randInt(max int) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))
	return int(n.Int64())
}

func randFloat() float64 {
	n, _ := rand.Int(rand.Reader, big.NewInt(1000))
	return float64(n.Int64()) / 1000.0
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
