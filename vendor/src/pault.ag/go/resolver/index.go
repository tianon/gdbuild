package resolver

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"pault.ag/go/debian/control"
	"pault.ag/go/debian/dependency"
	"pault.ag/go/debian/version"
)

type Canidates map[string][]control.BinaryIndex

func (can *Canidates) AppendBinaryIndexReader(in io.Reader) error {
	reader := bufio.NewReader(in)
	index, err := control.ParseBinaryIndex(reader)
	if err != nil {
		return err
	}
	can.AppendBinaryIndex(index)
	return nil
}

func (can *Canidates) AppendBinaryIndex(index []control.BinaryIndex) {
	for _, entry := range index {
		(*can)[entry.Package] = append((*can)[entry.Package], entry)
	}
}

func NewCanidates(index []control.BinaryIndex) Canidates {
	ret := Canidates{}
	ret.AppendBinaryIndex(index)
	return ret
}

func ReadFromBinaryIndex(in io.Reader) (*Canidates, error) {
	reader := bufio.NewReader(in)
	index, err := control.ParseBinaryIndex(reader)
	if err != nil {
		return nil, err
	}
	can := NewCanidates(index)
	return &can, nil
}

func (can Canidates) ExplainSatisfiesBuildDepends(arch dependency.Arch, depends dependency.Dependency) (bool, string) {
	for _, possi := range depends.GetPossibilities(arch) {
		can, why := can.ExplainSatisfies(arch, possi)
		if !can {
			return false, fmt.Sprintf("Possi %s can't be satisfied - %s", possi.Name, why)
		}
	}
	return true, "All relations are a go"
}

func (can Canidates) SatisfiesBuildDepends(arch dependency.Arch, depends dependency.Dependency) bool {
	ret, _ := can.ExplainSatisfiesBuildDepends(arch, depends)
	return ret
}

func (can Canidates) Satisfies(arch dependency.Arch, possi dependency.Possibility) bool {
	ret, _ := can.ExplainSatisfies(arch, possi)
	return ret
}

func (can Canidates) ExplainSatisfies(arch dependency.Arch, possi dependency.Possibility) (bool, string) {
	entries, ok := can[possi.Name]
	if !ok { // no known entries in the Index
		return false, fmt.Sprintf("Totally unknown package: %s", possi.Name)
	}

	if possi.Arch != nil {
		satisfied := false
		for _, installable := range entries {
			if installable.Architecture.Is(possi.Arch) {
				satisfied = true
				break
			}
		}
		if !satisfied {
			return false, fmt.Sprintf(
				"Relation depends on multiarch arch %s-%s-%s. Not found",
				possi.Arch.ABI,
				possi.Arch.OS,
				possi.Arch.CPU,
			)
		}
	}

	if possi.Version == nil {
		return true, "Relation exists, no version constraint"
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
			return false, "Unknown operator D:" // XXX: WHAT THE SHIT
		}

		if satisfied {
			return true, "Relation exists with a satisfied version constraint"
		}
	}

	return false, fmt.Sprintf(
		"%s is version constrainted %s %s. Valid options: %s",
		possi.Name,
		vr.Operator,
		vr.Number,
		strings.Join(seenRealtions, ", "),
	)
}
