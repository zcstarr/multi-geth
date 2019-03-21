// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package ethash

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

func init() {
	rand.Seed(time.Now().Unix())
}

type diffTest struct {
	ParentTimestamp    uint64
	ParentDifficulty   *big.Int
	CurrentTimestamp   uint64
	CurrentBlocknumber *big.Int
	CurrentDifficulty  *big.Int
	ParentUnclesHash   common.Hash
	ChainConfig        *params.ChainConfig
}

func (d *diffTest) UnmarshalJSON(b []byte) (err error) {
	var ext struct {
		ParentTimestamp    string
		ParentDifficulty   string
		CurrentTimestamp   string
		CurrentBlocknumber string
		CurrentDifficulty  string
		ChainConfig        *params.ChainConfig
	}
	if err := json.Unmarshal(b, &ext); err != nil {
		return err
	}

	d.ParentTimestamp = math.MustParseUint64(ext.ParentTimestamp)
	d.ParentDifficulty = math.MustParseBig256(ext.ParentDifficulty)
	d.CurrentTimestamp = math.MustParseUint64(ext.CurrentTimestamp)
	d.CurrentBlocknumber = math.MustParseBig256(ext.CurrentBlocknumber)
	d.CurrentDifficulty = math.MustParseBig256(ext.CurrentDifficulty)
	d.ChainConfig = ext.ChainConfig

	return nil
}

func TestCalcDifficulty(t *testing.T) {
	doTest := func(testfile string) {
		file, err := os.Open(testfile)
		if err != nil {
			t.Skip(err)
		}
		defer file.Close()

		tests := make(map[string]diffTest)
		err = json.NewDecoder(file).Decode(&tests)
		if err != nil {
			t.Fatal(err)
		}

		for name, test := range tests {
			config := test.ChainConfig
			if config == nil {
				config = &params.ChainConfig{HomesteadBlock: big.NewInt(1150000)}
			}
			number := new(big.Int).Sub(test.CurrentBlocknumber, big.NewInt(1))
			diff := CalcDifficulty(config, test.CurrentTimestamp, &types.Header{
				Number:     number,
				Time:       new(big.Int).SetUint64(test.ParentTimestamp),
				Difficulty: test.ParentDifficulty,
			})
			if diff.Cmp(test.CurrentDifficulty) != 0 {
				t.Error(name, "failed. Expected", test.CurrentDifficulty, "and calculated", diff, "test:", test)
			}
		}
	}
	for _, s := range []string{
		filepath.Join("..", "..", "tests", "testdata", "BasicTests", "difficulty.json"),
		filepath.Join("..", "..", "tests", "testdata", "BasicTests", "difficulty2.json"),
	} {
		doTest(s)
	}
}

// TestGenTestsCalcDifficulties is just an adhoc generator function to create JSON tests
// for a fuzzy-ish set of params beyond what is tested in the test above.
func TestGenTestsCalcDifficulties(t *testing.T) {
	chains := []*params.ChainConfig{
		params.TestChainConfig,
		params.MainnetChainConfig,
		params.ClassicChainConfig,
	}

	type testcaseS struct {
		ParentTimestamp    string
		CurrentTimestamp   string
		ParentDifficulty   string
		CurrentDifficulty  string
		ParentUnclesHash   string
		CurrentBlockNumber string
		ChainConfig        *params.ChainConfig
	}
	t2s := func(tc *diffTest) *testcaseS {
		return &testcaseS{
			ParentTimestamp:    fmt.Sprintf("%d", tc.ParentTimestamp),
			CurrentTimestamp:   fmt.Sprintf("%d", tc.CurrentTimestamp),
			ParentDifficulty:   fmt.Sprintf("%v", tc.ParentDifficulty),
			CurrentDifficulty:  fmt.Sprintf("%v", tc.CurrentDifficulty),
			ParentUnclesHash:   tc.ParentUnclesHash.String(),
			CurrentBlockNumber: fmt.Sprintf("%v", tc.CurrentBlocknumber),
			ChainConfig:        tc.ChainConfig,
		}
	}

	withSurroundingNumbers := func(edges []*big.Int) []*big.Int {
		var out []*big.Int
		for i := range edges {
			e := edges[i]
			if e == nil {
				continue
			}
			out = append(out, e)
			out = append(out, big.NewInt(0).Add(e, big.NewInt(1)))
			if e.Cmp(big.NewInt(0)) > 0 {
				out = append(out, new(big.Int).Sub(e, big.NewInt(1)))
			}
		}
		return out
	}
	difficultyInterestingForks := func(c *params.ChainConfig) []*big.Int {
		return []*big.Int{
			c.HomesteadBlock,
			c.EIP158Block,
			c.EIP100FBlock,
			c.EIP649FBlock,
			c.ECIP1010PauseBlock,
			c.ByzantiumBlock,
			c.DisposalBlock,
			c.EIP1234FBlock,
		}
	}

	maxTime := int32(999999999)
	maxTimeDelta := int32(42)
	genTestScene := func(c *params.ChainConfig, bn *big.Int) *diffTest {
		pt := rand.Int31n(maxTime)
		ct := pt + int32(rand.Int31n(maxTimeDelta))
		tc := &diffTest{
			ChainConfig:        c,
			ParentTimestamp:    uint64(pt),
			CurrentTimestamp:   uint64(ct),
			CurrentBlocknumber: bn,
			ParentDifficulty:   big.NewInt(0).SetUint64(uint64(rand.Int31n(maxTime))),
			ParentUnclesHash:   common.HexToHash("0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347"),
		}
		tc.CurrentDifficulty = CalcDifficulty(c, tc.CurrentTimestamp, &types.Header{
			Number:     big.NewInt(0).Sub(bn, big.NewInt(1)),
			Time:       new(big.Int).SetUint64(tc.ParentTimestamp),
			Difficulty: tc.ParentDifficulty,
		})

		return tc
	}

	var scenes []*diffTest
	for _, c := range chains {
		blocks := withSurroundingNumbers(difficultyInterestingForks(c))
		blocks = append(blocks, []*big.Int{big.NewInt(4200000), big.NewInt(9999999), big.NewInt(10000000)}...)
		t.Log("blocks", blocks)
		for _, b := range blocks {
			if b.Sign() == 0 {
				continue
			}
			s := genTestScene(c, b)
			scenes = append(scenes, s)
		}
	}

	var testdata = make(map[string]*testcaseS)
	for i, c := range scenes {
		testdata[strconv.Itoa(i)] = t2s(c)
	}

	b, err := json.MarshalIndent(testdata, "", "    ")
	if err != nil {
		t.Fatal(err)
	}

	// write := false
	write := true

	if write {
		file := filepath.Join("..", "..", "tests", "testdata", "BasicTests", "difficulty2.json")
		err = ioutil.WriteFile(file, b, os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}
	}
}
