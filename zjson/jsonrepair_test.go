package zjson_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zjson"
)

func TestRepair(t *testing.T) {
	tt := zlsgo.NewTest(t)
	tests := []struct {
		expected interface{}
		name     string
		input    string
		wantErr  bool
	}{
		{
			name:     "正常的JSON",
			input:    `{"name":"张三","age":30}`,
			expected: map[string]interface{}{"name": "张三", "age": float64(30)},
			wantErr:  false,
		},
		{
			name:     "缺少引号的键",
			input:    `{name:"张三","age":30}`,
			expected: map[string]interface{}{"name": "张三", "age": float64(30)},
			wantErr:  false,
		},
		{
			name:     "缺少逗号",
			input:    `{"name":"张三" "age":30}`,
			expected: map[string]interface{}{"name": "张三", "age": float64(30)},
			wantErr:  false,
		},
		{
			name:     "多余的逗号",
			input:    `{"name":"张三", "age":30,}`,
			expected: map[string]interface{}{"name": "张三", "age": float64(30)},
			wantErr:  false,
		},
		{
			name:     "单引号",
			input:    `{'name':'张三', 'age':30}`,
			expected: map[string]interface{}{"name": "张三", "age": float64(30)},
			wantErr:  false,
		},
		{
			name:     "布尔值和null",
			input:    `{"isActive":true, "deleted":False, "data":null}`,
			expected: map[string]interface{}{"isActive": true, "deleted": false, "data": nil},
			wantErr:  false,
		},
		{
			name: "JavaScript注释",
			input: `{
				// 这是一条注释
				"name": "张三", /* 这是另一条注释 */
				"age": 30
			}`,
			expected: map[string]interface{}{"name": "张三", "age": float64(30)},
			wantErr:  false,
		},
		{
			name:     "数组缺少逗号",
			input:    `{"items":[1 2 3]}`,
			expected: map[string]interface{}{"items": []interface{}{float64(1), float64(2), float64(3)}},
			wantErr:  false,
		},
		{
			name:  "嵌套对象和数组",
			input: `{"user":{"name":"张三","skills":["Go","Python" "Java"]}}`,
			expected: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "张三",
					"skills": []interface{}{
						"Go", "Python", "Java",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Markdown代码块格式的JSON",
			input: "```json\n" +
				"{\n" +
				"	\"name\": \"John\",\n" +
				"	\"age\": 30,\n" +
				"	\"isMarried\": false\n" +
				"}\n" +
				"```",
			expected: map[string]interface{}{"name": "John", "age": float64(30), "isMarried": false},
			wantErr:  false,
		},
		{
			name:     "空数组",
			input:    "[]",
			expected: []interface{}{},
			wantErr:  false,
		},
		{
			name:     "空对象带空格",
			input:    "   {  }   ",
			expected: map[string]interface{}{},
			wantErr:  false,
		},
		{
			name:     "大写的布尔值和null",
			input:    "{\"key\": TRUE, \"key2\": FALSE, \"key3\": Null}",
			expected: map[string]interface{}{"key": true, "key2": false, "key3": nil},
			wantErr:  false,
		},
		{
			name:     "混合引号和未引用键",
			input:    "{'key': 'string', 'key2': false, \"key3\": null, \"key4\": unquoted}",
			expected: map[string]interface{}{"key": "string", "key2": false, "key3": nil, "key4": "unquoted"},
			wantErr:  false,
		},
		{
			name:     "未闭合的对象",
			input:    "{\"name\": \"John\", \"age\": 30, \"city\": \"New York\"",
			expected: map[string]interface{}{"name": "John", "age": float64(30), "city": "New York"},
			wantErr:  false,
		},
		{
			name:     "未引用的对象值",
			input:    "{\"name\": \"John\", \"age\": 30, \"city\": New York}",
			expected: map[string]interface{}{"name": "John", "age": float64(30), "city": "New York"},
			wantErr:  false,
		},
		{
			name:     "未引用的对象键",
			input:    "{\"name\": John, \"age\": 30, \"city\": \"New York\"}",
			expected: map[string]interface{}{"name": "John", "age": float64(30), "city": "New York"},
			wantErr:  false,
		},
		{
			name:     "未闭合的数组",
			input:    "[1, 2, 3,",
			expected: []interface{}{float64(1), float64(2), float64(3)},
			wantErr:  false,
		},
		{
			name:     "未闭合的嵌套数组",
			input:    "{\"employees\":[\"John\", \"Anna\",",
			expected: map[string]interface{}{"employees": []interface{}{"John", "Anna"}},
			wantErr:  false,
		},
		{
			name:     "只有左括号的对象",
			input:    "{",
			expected: map[string]interface{}{},
			wantErr:  false,
		},
		{
			name:     "只有左括号的数组",
			input:    "[",
			expected: []interface{}{},
			wantErr:  false,
		},
		{
			name:     "带换行的嵌套数组",
			input:    "[[1\n\n]",
			expected: []interface{}{[]interface{}{float64(1)}},
			wantErr:  false,
		},
		{
			name:     "空键",
			input:    "{\"\":true, \"key2\":\"value2\"}",
			expected: map[string]interface{}{"": true, "key2": "value2"},
			wantErr:  false,
		},
		{
			name:     "转义单引号",
			input:    "{\"text\": \"The quick brown fox won\\'t jump\"}",
			expected: map[string]interface{}{"text": "The quick brown fox won't jump"},
			wantErr:  false,
		},
		{
			name:     "HTML标签",
			input:    "{\"real_content\": \"Some string: Some other string Some string <a href=\\\"https://domain.com\\\">Some link</a>\"}",
			expected: map[string]interface{}{"real_content": "Some string: Some other string Some string <a href=\"https://domain.com\">Some link</a>"},
			wantErr:  false,
		},
		{
			name:     "带转义的键",
			input:    "{\"key\\_1\": \"value\"}",
			expected: map[string]interface{}{"key_1": "value"},
			wantErr:  false,
		},
		{
			name:     "多重引号",
			input:    "{\"answer\":[{\"traits\":\"Female aged 60+\",\"answer1\":\"5\"}]}",
			expected: map[string]interface{}{"answer": []interface{}{map[string]interface{}{"traits": "Female aged 60+", "answer1": "5"}}},
			wantErr:  false,
		},
		{
			name: "复杂嵌套结构",
			input: `{
				  "resourceType": "Bundle",
				  "id": "1",
				  "type": "collection",
				  "entry": [
					{
					  "resource": {
						"resourceType": "Patient",
						"id": "1",
						"name": [
						  {"use": "official", "family": "Corwin", "given": ["Keisha", "Sunny"], "prefix": ["Mrs."]},
						  {"use": "maiden", "family": "Goodwin", "given": ["Keisha", "Sunny"], "prefix": ["Mrs."]}
						]
					  }
					}
				  ]
				}`,
			expected: map[string]interface{}{
				"resourceType": "Bundle",
				"id":           "1",
				"type":         "collection",
				"entry": []interface{}{
					map[string]interface{}{
						"resource": map[string]interface{}{
							"resourceType": "Patient",
							"id":           "1",
							"name": []interface{}{
								map[string]interface{}{
									"use":    "official",
									"family": "Corwin",
									"given":  []interface{}{"Keisha", "Sunny"},
									"prefix": []interface{}{"Mrs."},
								},
								map[string]interface{}{
									"use":    "maiden",
									"family": "Goodwin",
									"given":  []interface{}{"Keisha", "Sunny"},
									"prefix": []interface{}{"Mrs."},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:     "HTML内容带引号",
			input:    "{\"html\": \"<h3 id=\\\"aaa\\\">Waarom meer dan 200 Technical Experts - \\\"Passie voor techniek\\\"?</h3>\"}",
			expected: map[string]interface{}{"html": "<h3 id=\"aaa\">Waarom meer dan 200 Technical Experts - \"Passie voor techniek\"?</h3>"},
			wantErr:  false,
		},
		{
			name:     "小数点开头的数字",
			input:    "{\"key\": .25}",
			expected: map[string]interface{}{"key": 0.25},
			wantErr:  false,
		},
		{
			name:     "不匹配字面量",
			input:    "{ \"words\": abcdef\", \"numbers\": 12345\", \"words2\": ghijkl\" }",
			expected: map[string]interface{}{"words": "abcdef", "numbers": float64(12345), "words2": "ghijkl"},
			wantErr:  false,
		},
		{
			name:     "冲突的双引号",
			input:    "[{\"Master\":\"господин\"}]",
			expected: []interface{}{map[string]interface{}{"Master": "господин"}},
			wantErr:  false,
		},
		{
			name:     "空JSON",
			input:    ``,
			expected: `""`,
			wantErr:  false,
		},
		{
			name:     "完全无效的JSON",
			input:    `这不是JSON`,
			expected: "\"这不是JSON\"",
			wantErr:  false,
		},
	}

	for _, ts := range tests {
		t.Run(ts.name, func(t *testing.T) {
			got, err := zjson.Repair(ts.input)

			tt.NoError(err)

			if got == "" && ts.expected == "" {
				return
			}

			var gotObj interface{}
			if err := json.Unmarshal([]byte(got), &gotObj); err != nil {
				t.Errorf("无法解析修复后的 JSON: %v", err)
				return
			}

			if expectedStr, ok := ts.expected.(string); ok {
				tt.Equal(got, expectedStr)
				return
			}

			tt.Equal(gotObj, ts.expected)
		})
	}
}

func jsonStringsEqual(jsonStr1, jsonStr2 string) bool {
	var jsonObj1, jsonObj2 interface{}

	if err := json.Unmarshal([]byte(jsonStr1), &jsonObj1); err != nil {
		return false
	}

	if err := json.Unmarshal([]byte(jsonStr2), &jsonObj2); err != nil {
		return false
	}

	return reflect.DeepEqual(jsonObj1, jsonObj2)
}

func TestRepairWithReferenceTestCases(t *testing.T) {
	tt := zlsgo.NewTest(t)
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{
			name: "基本JSON对象",
			in: `
				{
					"name": "John",
					"age": 30,
					"isMarried": false
				}`,
			want: `{"name":"John","age":30,"isMarried":false}`,
		},
		{
			name: "Markdown代码块",
			in: "```json\n" +
				"{\n" +
				"	\"name\": \"John\",\n" +
				"	\"age\": 30,\n" +
				"	\"isMarried\": false\n" +
				"}\n" +
				"```",
			want: `{"name":"John","age":30,"isMarried":false}`,
		},
		{
			name: "空数组",
			in:   "[]",
			want: `[]`,
		},
		{
			name: "空对象带空格",
			in:   "   {  }   ",
			want: `{}`,
		},
		{
			name: "空字符串",
			in:   `"`,
			want: `""`,
		},
		{
			name: "只有换行符",
			in:   "\n",
			want: `""`,
		},
		{
			name: "标准的属性",
			in:   `  {"key": true, "key2": false, "key3": null}`,
			want: `{"key":true,"key2":false,"key3":null}`,
		},
		{
			name: "大写布尔值",
			in:   "{\"key\": TRUE, \"key2\": FALSE, \"key3\": Null } ",
			want: `{"key":true,"key2":false,"key3":null}`,
		},
		{
			name: "未闭合的大写布尔值",
			in:   "{\"key\": TRUE, \"key2\": FALSE, \"key3\": Null  ",
			want: `{"key":true,"key2":false,"key3":null}`,
		},
		{
			name: "混合引号和未引用键",
			in:   "{'key': 'string', 'key2': false, \"key3\": null, \"key4\": unquoted}",
			want: `{"key":"string","key2":false,"key3":null,"key4":"unquoted"}`,
		},
		{
			name: "标准JSON",
			in:   `{"name": "John", "age": 30, "city": "New York"}`,
			want: `{"name":"John","age":30,"city":"New York"}`,
		},
		{
			name: "数组",
			in:   "[1, 2, 3, 4]",
			want: `[1,2,3,4]`,
		},
		{
			name: "未闭合的数组",
			in:   "[1, 2, 3, 4",
			want: `[1,2,3,4]`,
		},
		{
			name: "带空格的数组",
			in:   `{"employees":["John", "Anna", "Peter"]} `,
			want: `{"employees":["John","Anna","Peter"]}`,
		},
		{
			name: "未闭合的对象",
			in:   `{"name": "John", "age": 30, "city": "New York`,
			want: `{"name":"John","age":30,"city":"New York"}`,
		},
		{
			name: "未引用的对象键",
			in:   `{"name": "John", "age": 30, city: "New York"}`,
			want: `{"name":"John","age":30,"city":"New York"}`,
		},
		{
			name: "未引用的对象值",
			in:   `{"name": "John", "age": 30, "city": New York}`,
			want: `{"name":"John","age":30,"city":"New York"}`,
		},
		{
			name: "未引用的键值",
			in:   `{"name": John, "age": 30, "city": "New York"}`,
			want: `{"name":"John","age":30,"city":"New York"}`,
		},
		{
			name: "未闭合的数组2",
			in:   `[1, 2, 3,`,
			want: `[1,2,3]`,
		},
		{
			name: "未闭合的嵌套数组",
			in:   `{"employees":["John", "Anna",`,
			want: `{"employees":["John","Anna"]}`,
		},
		{
			name: "空格字符串",
			in:   " ",
			want: `""`,
		},
		{
			name: "只有左括号的数组",
			in:   "[",
			want: "[]",
		},
		{
			name: "只有右括号的数组",
			in:   "]",
			want: `"]"`,
		},
		{
			name: "带换行的嵌套数组",
			in:   "[[1\n\n]",
			want: "[[1]]",
		},
		{
			name: "只有左括号的对象",
			in:   "{",
			want: "{}",
		},
		{
			name: "只有右括号的对象",
			in:   "}",
			want: `"}"`,
		},
		{
			name:    "未闭合的对象键",
			in:      `{"`,
			want:    `{}`,
			wantErr: true,
		},
		{
			name:    "未闭合的数组值",
			in:      `["`,
			want:    `[]`,
			wantErr: true,
		},
		{
			name: "单引号的引号",
			in:   `'\"'`,
			want: `""`,
		},
		{
			name: "普通字符串",
			in:   "string",
			want: `"string"`,
		},
		{
			name:    "未闭合的花括号和方括号",
			in:      `{foo: [}`,
			want:    `{"foo":[]}`,
			wantErr: true,
		},
		{
			name: "键中包含冒号",
			in:   `{"key": "value:value"}`,
			want: `{"key":"value:value"}`,
		},
		{
			name: "未闭合的对象",
			in:   `{"name": "John", "age": 30, "city": "New`,
			want: `{"name":"John","age":30,"city":"New"}`,
		},
		{
			name: "未闭合的嵌套数组2",
			in:   `{"employees":["John", "Anna", "Peter`,
			want: `{"employees":["John","Anna","Peter"]}`,
		},
		{
			name: "完整的对象和数组",
			in:   `{"employees":["John", "Anna", "Peter"]}`,
			want: `{"employees":["John","Anna","Peter"]}`,
		},
		{
			name: "带逗号的文本",
			in:   `{"text": "The quick brown fox,"}`,
			want: `{"text":"The quick brown fox,"}`,
		},
		{
			name: "带转义单引号的文本",
			in:   `{"text": "The quick brown fox won\'t jump"}`,
			want: `{"text":"The quick brown fox won't jump"}`,
		},
		{
			name: "值中有多个冒号",
			in:   `{"value_1": "value_2": "data"}`,
			want: `{"value_1":"value_2"}`,
		},
		{
			name: "带标记的JSON",
			in:   `{"value_1": true, COMMENT "value_2": "data"}`,
			want: `{"value_1":true,"value_2":"data"}`,
		},
		{
			name: "带注释的JSON",
			in:   `{"value_1": true, /* value_2 */ "value_2": "data"}`,
			want: `{"value_1":true,"value_2":"data"}`,
		},
		{
			name: "带多余文本的JSON",
			in:   `{"value_1": true, SHOULD_NOT_EXIST "value_2": "data" AAAA }`,
			want: `{"value_1":true,"value_2":"data"}`,
		},
		{
			name: "空键",
			in:   `{"": true, "key2": "value2"}`,
			want: `{"":true,"key2":"value2"}`,
		},
		{
			name: "带前缀的JSON",
			in:   ` - { "test_key": ["test_value", "test_value2"] }`,
			want: `{"test_key":["test_value","test_value2"]}`,
		},
		{
			name: "复杂嵌套URL",
			in:   `{ "content": "[LINK]("https://google.com")" }`,
			want: `{"content":"[LINK](","https":""}`,
		},
		{
			name: "链接URI格式1",
			in:   `{ "content": "[LINK](" }`,
			want: `{"content":"[LINK]("}`,
		},
		{
			name: "链接URI格式2",
			in:   `{ "content": "[LINK](", "key": true }`,
			want: `{"content":"[LINK](","key":true}`,
		},
		{
			name: "带代码块的JSON",
			in: "```json\n" +
				"{\n" +
				"	\"key\": \"value\"\n" +
				"}\n" +
				"```",
			want: `{"key":"value"}`,
		},
		{
			name: "不规则的代码块",
			in:   "````{ \"key\": \"value\" }```",
			want: `{"key":"value"}`,
		},
		{
			name: "带HTML的JSON",
			in:   `{"real_content": "Some string: Some other string Some string <a href=\"https://domain.com\">Some  link</a>"}`,
			want: `{"real_content":"Some string: Some other string Some string <a href=\"https://domain.com\">Some  link</a>"}`,
		},
		{
			name: "带转义的键",
			in:   "{\"key\\_1\n\": \"value\"}",
			want: `{"key_1":"value"}`,
		},
		{
			name: "带Tab和转义的键",
			in:   "{\"key\t\\_\": \"value\"}",
			want: `{"key\t_":"value"}`,
		},
		{
			name:    "多重引号",
			in:      `{""answer"":[{""traits"":''Female aged 60+'',""answer1"":""5""}]}`,
			want:    `{"answer":[{"traits":"Female aged 60+","answer1":"5"}]}`,
			wantErr: true,
		},
		{
			name: "不匹配的字符串边界",
			in:   `{ "words": abcdef", "numbers": 12345", "words2": ghijkl" }`,
			want: `{"words":"abcdef","numbers":12345,"words2":"ghijkl"}`,
		},
		{
			name: "小数点开头的数字",
			in:   `{"key": .25}`,
			want: `{"key":0.25}`,
		},
		{
			name: "复杂的嵌套结构",
			in: `{
				  "resourceType": "Bundle",
				  "id": "1",
				  "type": "collection",
				  "entry": [
					{
					  "resource": {
						"resourceType": "Patient",
						"id": "1",
						"name": [
						  {"use": "official", "family": "Corwin", "given": ["Keisha", "Sunny"], "prefix": ["Mrs."]},
						  {"use": "maiden", "family": "Goodwin", "given": ["Keisha", "Sunny"], "prefix": ["Mrs."]}
						]
					  }
					}
				  ]
				}`,
			want: `{"resourceType":"Bundle","id":"1","type":"collection","entry":[{"resource":{"resourceType":"Patient","id":"1","name":[{"use":"official","family":"Corwin","given":["Keisha","Sunny"],"prefix":["Mrs."]},{"use":"maiden","family":"Goodwin","given":["Keisha","Sunny"],"prefix":["Mrs."]}]}}]}`,
		},
		{
			name: "HTML带引号",
			in:   `{"html": "<h3 id=\"aaa\">Waarom meer dan 200 Technical Experts - \"Passie voor techniek\"?</h3>"}`,
			want: `{"html":"<h3 id=\"aaa\">Waarom meer dan 200 Technical Experts - \"Passie voor techniek\"?</h3>"}`,
		},
		{
			name: "单引号评论",
			in:   `{  'reviews': [    {      'version': 'new',      'line': 1,      'severity': 'Minor',      'issue_type': 'Standard practice suggestion',      'issue': 'The merge request description is missing a link to the original issue or bug report.',      'suggestions': 'Add a link to the original issue or bug report in the *Issue* section.'    },    {      'version': 'new',      'line': 2,      'severity': 'Minor',      'issue_type': 'Standard practice suggestion',      'issue': 'The merge request description is missing a description of the critical issue or bug being addressed.',      'suggestions': 'Add a description of the critical issue or bug being addressed in the *Problem* section.'    } ]`,
			want: `{"reviews":[{"version":"new","line":1,"severity":"Minor","issue_type":"Standard practice suggestion","issue":"The merge request description is missing a link to the original issue or bug report.","suggestions":"Add a link to the original issue or bug report in the *Issue* section."},{"version":"new","line":2,"severity":"Minor","issue_type":"Standard practice suggestion","issue":"The merge request description is missing a description of the critical issue or bug being addressed.","suggestions":"Add a description of the critical issue or bug being addressed in the *Problem* section."}]}`,
		},
		{
			name: "尾部逗号",
			in:   `{"key":"",}`,
			want: `{"key":""}`,
		},
		{
			name: "多行代码块",
			in:   "```json{\"array_key\": [{\"item_key\": 1\n}], \"outer_key\": 2}```",
			want: `{"array_key":[{"item_key":1}],"outer_key":2}`,
		},
		{
			name: "特殊字符",
			in: `[
	{"Master""господин"}
	]`,
			want:    `[{"Master":"господин"}]`,
			wantErr: true,
		},
	}

	for _, ts := range tests {
		t.Run(ts.name, func(t *testing.T) {
			got, err := zjson.Repair(ts.in)
			if ts.wantErr {
				tt.EqualTrue(err != nil)
				return
			}
			tt.NoError(err, true)

			if !jsonStringsEqual(got, ts.want) {
				t.Errorf("%s: Repair() = %v, want %v", ts.name, got, ts.want)
			}
		})
	}
}

func TestRepairWithOptions(t *testing.T) {
	tt := zlsgo.NewTest(t)
	tests := []struct {
		expected interface{}
		options  *zjson.RepairOptions
		name     string
		input    string
		wantErr  bool
	}{
		{
			name:  "禁止单引号",
			input: `{'name':'张三'}`,
			options: &zjson.RepairOptions{
				AllowSingleQuotes:   false,
				AllowComments:       true,
				AllowTrailingCommas: true,
				AllowUnquotedKeys:   true,
			},
			wantErr: true,
		},
		{
			name:  "禁止未引用的键",
			input: `{name:"张三"}`,
			options: &zjson.RepairOptions{
				AllowSingleQuotes:   true,
				AllowComments:       true,
				AllowTrailingCommas: true,
				AllowUnquotedKeys:   false,
			},
			wantErr: true,
		},
		{
			name: "禁止注释",
			input: `{
				// 这是注释
				"name": "张三"
			}`,
			options: &zjson.RepairOptions{
				AllowSingleQuotes:   true,
				AllowComments:       false,
				AllowTrailingCommas: true,
				AllowUnquotedKeys:   true,
			},
			expected: map[string]interface{}{"name": "张三"},
			wantErr:  false,
		},
		{
			name:  "禁止尾随逗号",
			input: `{"name":"张三","items":[1,2,3,]}`,
			options: &zjson.RepairOptions{
				AllowSingleQuotes:   true,
				AllowComments:       true,
				AllowTrailingCommas: false,
				AllowUnquotedKeys:   true,
			},
			wantErr: true,
		},
	}

	for _, ts := range tests {
		t.Run(ts.name, func(t *testing.T) {
			got, err := zjson.Repair(ts.input, func(opts *zjson.RepairOptions) {
				*opts = *ts.options
			})
			if ts.wantErr {
				tt.EqualTrue(err != nil)
				return
			}
			tt.NoError(err, true)

			var gotObj interface{}
			if err := json.Unmarshal([]byte(got), &gotObj); err != nil {
				t.Errorf("无法解析修复后的 JSON: %v", err)
				return
			}

			if !reflect.DeepEqual(gotObj, ts.expected) {
				t.Errorf("RepairWithOptions() = %v, want %v", gotObj, ts.expected)
			}
		})
	}
}

func BenchmarkRepair(b *testing.B) {
	samples := []string{
		`{"name":"张三","age":30,"active":true}`,
		`{
			// 这是注释
			"name": "张三",
			"age": 30 /* 内联注释 */
		}`,
		`{name: "张三", 'age': 30, items: [1 2 3]}`,
	}

	for i, sample := range samples {
		b.Run(fmt.Sprintf("Sample%d", i+1), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				zjson.Repair(sample)
			}
		})
	}
}

func BenchmarkRepairWithOptions(b *testing.B) {
	sample := `{
		// 这是注释
		name: "张三",
		'age': 30,
		items: [1 2 3,]
	}`

	options := []*zjson.RepairOptions{
		{AllowComments: false, AllowSingleQuotes: true, AllowTrailingCommas: true, AllowUnquotedKeys: true},
		{AllowComments: true, AllowSingleQuotes: false, AllowTrailingCommas: true, AllowUnquotedKeys: true},
		{AllowComments: true, AllowSingleQuotes: true, AllowTrailingCommas: false, AllowUnquotedKeys: true},
		{AllowComments: true, AllowSingleQuotes: true, AllowTrailingCommas: true, AllowUnquotedKeys: false},
	}

	for i, opt := range options {
		b.Run(fmt.Sprintf("Options%d", i+1), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				zjson.Repair(sample, func(opts *zjson.RepairOptions) {
					*opts = *opt
				})
			}
		})
	}
}

func BenchmarkRepairComplex(b *testing.B) {
	samples := []string{
		`{
			"resourceType": "Bundle",
			"id": "1",
			"type": "collection",
			"entry": [
			  {
				"resource": {
				  "resourceType": "Patient",
				  "id": "1",
				  "name": [
					{"use": "official", "family": "Corwin", "given": ["Keisha", "Sunny"], "prefix": ["Mrs."]},
					{"use": "maiden", "family": "Goodwin", "given": ["Keisha", "Sunny"], "prefix": ["Mrs."]}
				  ]
				}
			  }
			]
		}`,
		`{
			name: 'John', 
			"age": 30, 
			'hobbies': ["reading", "swimming" coding], 
			"address":{
				city: "New York"
				"country": USA
				zip: "10001",
			},
			"isMarried": True,
			"children": null,
		}`,
		`{
			// 这是用户信息
			'user': {
				"name": "张三", /* 中文名字 */
				'age': 30, // 年龄
				"skills": ["Java", 'Python' "Go"],
				"contact": {
					"email": "zhangsan@example.com",
					phone: "12345678"
				}
			}
		}`,
	}

	for i, sample := range samples {
		b.Run(fmt.Sprintf("ComplexSample%d", i+1), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				zjson.Repair(sample)
			}
		})
	}
}

func TestJSONRepairEdgeCases(t *testing.T) {
	tt := zlsgo.NewTest(t)

    tests := []zlsgo.ErrorTestCase{
		{
			Name:     "missing closing brace",
			Input:    `{"a":1`,
			Expected: `{"a":1}`,
			WantErr:  false,
		},
		{
			Name:     "missing closing bracket",
			Input:    `[1,2,3`,
			Expected: `[1,2,3]`,
			WantErr:  false,
		},
		{
			Name:     "single quotes",
			Input:    `{'a':1}`,
			Expected: `{"a":1}`,
			WantErr:  false,
		},
		{
			Name:     "unquoted keys",
			Input:    `{a:1,b:2}`,
			Expected: `{"a":1,"b":2}`,
			WantErr:  false,
		},
	}

    tt.RunErrorTests(tests, func(input interface{}) (interface{}, error) {
        return zjson.Repair(input.(string))
	})
}
