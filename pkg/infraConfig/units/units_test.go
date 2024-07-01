/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package units

import (
	"fmt"
	"testing"
)

// todo: add more test cases
func TestParseQuantityString(t *testing.T) {
	memLimit := "01.400Gi"
	pos, val, num, denom, suf, err := ParseQuantityString(memLimit)
	fmt.Println("pos: ", pos)
	fmt.Println("val: ", val)
	fmt.Println("num: ", num)
	fmt.Println("denom: ", denom)
	fmt.Println("suf: ", suf)
	fmt.Println("err: ", err)

}
