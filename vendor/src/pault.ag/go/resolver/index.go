package resolver // import "pault.ag/go/resolver"

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"pault.ag/go/debian/control"
	"pault.ag/go/debian/dependency"
	"pault.ag/go/debian/version"
)

type Candidates map[string][]control.BinaryIndex

func (can *Candidates) AppendBinaryIndexReader(in io.Reader) error {
	reader := bufio.NewReader(in)
	index, err := control.ParseBinaryIndex(reader)
	if err != nil {
		return err
	}
	can.AppendBinaryIndex(index)
	return nil
}

func (can *Candidates) AppendBinaryIndex(index []control.BinaryIndex) {
	for _, entry := range index {
		(*can)[entry.Package] = append((*can)[entry.Package], entry)
	}
}

func NewCandidates(index []control.BinaryIndex) Candidates {
	ret := Candidates{}
	ret.AppendBinaryIndex(index)
	return ret
}

func ReadFromBinaryIndex(in io.Reader) (*Candidates, error) {
	reader := bufio.NewReader(in)
	index, err := control.ParseBinaryIndex(reader)
	if err != nil {
		return nil, err
	}
	can := NewCandidates(index)
	return &can, nil
}

func (can Candidates) ExplainSatisfiesBuildDepends(arch dependency.Arch, depends dependency.Dependency) (bool, string) {
	for _, possi := range depends.GetPossibilities(arch) {
		can, why, _ := can.ExplainSatisfies(arch, possi)
		if !can {
			return false, fmt.Sprintf("Possi %s can't be satisfied - %s", possi.Name, why)
		}
	}
	return true, "All relations are a go"
}

func (can Candidates) SatisfiesBuildDepends(arch dependency.Arch, depends dependency.Dependency) bool {
	ret, _ := can.ExplainSatisfiesBuildDepends(arch, depends)
	return ret
}

func (can Candidates) Satisfies(arch dependency.Arch, possi dependency.Possibility) bool {
	ret, _, _ := can.ExplainSatisfies(arch, possi)
	return ret
}

func (can Candidates) ExplainSatisfies(arch dependency.Arch, possi dependency.Possibility) (bool, string, []control.BinaryIndex) {
	entries, ok := can[possi.Name]
	if !ok { // no known entries in the Index
		return false, fmt.Sprintf("Totally unknown package: %s", possi.Name), nil
	}

	if possi.Arch != nil {
		satisfied := false
		archEntries := []control.BinaryIndex{}
		for _, installable := range entries {
			if installable.Architecture.Is(possi.Arch) {
				archEntries = append(archEntries, installable)
				satisfied = true
			}
		}
		if !satisfied {
			return false, fmt.Sprintf(
				"Relation depends on multiarch arch %s-%s-%s. Not found",
				possi.Arch.ABI,
				possi.Arch.OS,
				possi.Arch.CPU,
			), nil
		}
		entries = archEntries
	}

	if possi.Version == nil {
		return true, "Relation exists, no version constraint", entries
	}

	// OK, so we have to play with versions now.
	vr := *possi.Version
	relatioNumber, _ := version.Parse(vr.Number)
	satisfied := false
	seenRealtions := []string{}

	for _, installable := range entries {
		q := version.Compare(installable.Version, relatioNumber)
		seenRealtions = append(seenRealtions, installable.Version.String())

		switch vr.Operator {
		case ">=":
			satisfied = q >= 0
		case "<=":
			satisfied = q <= 0
		case ">>":
			satisfied = q > 0
		case "<<":
			satisfied = q < 0
		case "=":
			satisfied = q == 0
		default:
			return false, "Unknown operator D:", nil // XXX: WHAT THE SHIT
		}

		if satisfied {
			return true, "Relation exists with a satisfied version constraint", []control.BinaryIndex{installable} // TODO gather the full list of version-constrained satisfiers
		}
	}

	return false, fmt.Sprintf(
		"%s is version constrainted %s %s. Valid options: %s",
		possi.Name,
		vr.Operator,
		vr.Number,
		strings.Join(seenRealtions, ", "),
	), nil
}
