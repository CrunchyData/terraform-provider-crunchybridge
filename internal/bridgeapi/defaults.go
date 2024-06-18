/*
Copyright 2022 Crunchy Data Solutions, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package bridgeapi

import (
	"fmt"
	"math/rand" // Should be sufficient for random names over crypto/rand
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

// returns a randomized cluster name in the form of
// adverb-adjective-animal-dd (d = digit from 1-8)
func DefaultClusterName() string {
	entropy := rand.Uint32()

	// Binary form, 32 bits, 4.29 Bil combos:
	// |   8    |   8    |  10 |   3    |   3    |
	// | adverb | animal | adj | digit1 | digit2 |

	// 8 bits for adverb
	ixAdverb := (entropy & 0xFF000000) >> 24
	// 8 bits for animal
	ixAnimal := (entropy & 0x00FF0000) >> 16
	// 10 bits for adj
	ixAdj := (entropy & 0x0000FFC0) >> 6
	// 3 bits for digit 1
	digit1 := 1 + (entropy&0x00000038)>>3
	// 3 bits for digit 2
	digit2 := 1 + (entropy & 0x00000007)

	return fmt.Sprintf("%s-%s-%s-%d%d", adverbs[ixAdverb], adjectives[ixAdj], animals[ixAnimal], digit1, digit2)
}
