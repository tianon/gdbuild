/* {{{ Copyright (c) Paul R. Tagliamonte <paultag@gmail.com>, 2015
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE. }}} */

package topsort_test

import (
	"log"
	"testing"

	"pault.ag/go/topsort"
)

// Test Helpers {{{

func isok(t *testing.T, err error) {
	if err != nil {
		log.Printf("Error! Error is not nil! %s\n", err)
		t.FailNow()
	}
}

func notok(t *testing.T, err error) {
	if err == nil {
		log.Printf("Error! Error is nil!\n")
		t.FailNow()
	}
}

func assert(t *testing.T, expr bool) {
	if !expr {
		log.Printf("Assertion failed!")
		t.FailNow()
	}
}

// }}}

func TestTopsortEasy(t *testing.T) {
	network := topsort.NewNetwork()
	network.AddNode("foo", nil)
	network.AddNode("bar", nil)
	network.AddEdge("foo", "bar")
	series, err := network.Sort()
	isok(t, err)
	assert(t, len(series) == 2)
}

func TestTopsortCycle(t *testing.T) {
	network := topsort.NewNetwork()
	network.AddNode("foo", nil)
	network.AddNode("bar", nil)
	network.AddEdge("foo", "bar")
	network.AddEdge("bar", "foo")
	_, err := network.Sort()
	notok(t, err)

}

func TestTopsortLongCycle(t *testing.T) {
	network := topsort.NewNetwork()

	network.AddNode("A", nil)
	network.AddNode("B", nil)
	network.AddNode("C", nil)
	network.AddNode("D", nil)
	network.AddNode("E", nil)
	network.AddNode("F", nil)
	network.AddNode("G", nil)
	network.AddNode("H", nil)
	network.AddNode("I", nil)

	network.AddEdge("A", "B")
	network.AddEdge("B", "C")
	network.AddEdge("C", "D")
	network.AddEdge("D", "E")
	network.AddEdge("E", "F")
	network.AddEdge("F", "G")
	network.AddEdge("G", "H")
	network.AddEdge("H", "I")
	network.AddEdge("I", "A")

	_, err := network.Sort()
	notok(t, err)
}

func TestTopsortLong(t *testing.T) {
	/*
		A --> B -> C -> D -> E
		 \    ^
		  \-> F
	*/
	network := topsort.NewNetwork()

	network.AddNode("A", nil)
	network.AddNode("B", nil)
	network.AddNode("C", nil)
	network.AddNode("D", nil)
	network.AddNode("E", nil)
	network.AddNode("F", nil)

	network.AddEdge("A", "B")
	network.AddEdge("A", "F")
	network.AddEdge("F", "B")
	network.AddEdge("B", "C")
	network.AddEdge("C", "D")
	network.AddEdge("D", "E")

	series, err := network.Sort()
	isok(t, err)
	assert(t, len(series) == 6)

	assert(t, series[0].Name == "A")
	assert(t, series[1].Name == "F")
	assert(t, series[2].Name == "B")
	assert(t, series[3].Name == "C")
}

// vim: foldmethod=marker
