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
	"math/big"
	"math/rand"
	"os"
	"path/filepath"
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
}

func (d *diffTest) UnmarshalJSON(b []byte) (err error) {
	var ext struct {
		ParentTimestamp    string
		ParentDifficulty   string
		CurrentTimestamp   string
		CurrentBlocknumber string
		CurrentDifficulty  string
	}
	if err := json.Unmarshal(b, &ext); err != nil {
		return err
	}

	d.ParentTimestamp = math.MustParseUint64(ext.ParentTimestamp)
	d.ParentDifficulty = math.MustParseBig256(ext.ParentDifficulty)
	d.CurrentTimestamp = math.MustParseUint64(ext.CurrentTimestamp)
	d.CurrentBlocknumber = math.MustParseBig256(ext.CurrentBlocknumber)
	d.CurrentDifficulty = math.MustParseBig256(ext.CurrentDifficulty)

	return nil
}

func TestCalcDifficulty(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "tests", "testdata", "BasicTests", "difficulty.json"))
	if err != nil {
		t.Skip(err)
	}
	defer file.Close()

	tests := make(map[string]diffTest)
	err = json.NewDecoder(file).Decode(&tests)
	if err != nil {
		t.Fatal(err)
	}

	config := &params.ChainConfig{HomesteadBlock: big.NewInt(1150000)}

	for name, test := range tests {
		number := new(big.Int).Sub(test.CurrentBlocknumber, big.NewInt(1))
		diff := CalcDifficulty(config, test.CurrentTimestamp, &types.Header{
			Number:     number,
			Time:       new(big.Int).SetUint64(test.ParentTimestamp),
			Difficulty: test.ParentDifficulty,
		})
		if diff.Cmp(test.CurrentDifficulty) != 0 {
			t.Error(name, "failed. Expected", test.CurrentDifficulty, "and calculated", diff)
		}
	}
}

// TestGenTestsCalcDifficulties is just an adhoc generator function to create JSON tests
// for a fuzzy-ish set of params beyond what is tested in the test above.
func TestGenTestsCalcDifficulties(t *testing.T) {
	// VARS:
	// chain config
	// parent timestamp
	// parent diff
	// parent uncles
	// current timestamp
	// current blocknumber
	//
	// OUT:
	// current difficulty

	chains := []*params.ChainConfig{
		params.TestChainConfig,
		params.MainnetChainConfig,
		params.ClassicChainConfig,
	}

	type testcase struct {
		parentTimestamp    uint64
		currentTimestamp   uint64
		parentDifficulty   *big.Int
		currentDifficulty  *big.Int
		parentUnclesHash   common.Hash
		currentBlockNumber uint64
	}

	withSurroundingNumbers := func(edges []*big.Int) []*big.Int {
		var out []*big.Int
		for i := range edges {
			e := edges[i]
			out = append(out, e)
			out = append(out, new(big.Int).Add(e, big.NewInt(1)))
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
			c.
		}
	}

	genTestScene := func(c *params.ChainConfig, bn *big.Int) *testcase {
		
	}

}
