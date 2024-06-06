package tests

import (
	"testing"

	fileeditor "github.com/Asiandayboy/CLITextEditor/fileEditor"
)

func TestCalcBufferLineFromACY(t *testing.T) {
	tests := []struct {
		name         string
		cursorY      int
		mappedBuffer []int
		expected     int
	}{
		{
			name:         "Test 1",
			cursorY:      10,
			mappedBuffer: []int{2, 4, 5, 7, 8, 12, 13},
			expected:     5,
		},
		{
			name:         "Test 2",
			cursorY:      3,
			mappedBuffer: []int{2, 4, 5, 7, 8, 12, 13},
			expected:     1,
		},
		{
			name:         "Test 3",
			cursorY:      1,
			mappedBuffer: []int{1, 2, 3, 4, 5, 8, 9},
			expected:     0,
		},
		{
			name:         "Test 4",
			cursorY:      7,
			mappedBuffer: []int{1, 2, 3, 4, 5, 8, 9},
			expected:     5,
		},
		{
			name:         "Test 5",
			cursorY:      8,
			mappedBuffer: []int{1, 2, 3, 4, 5, 8, 9},
			expected:     5,
		},
		{
			name:         "Test 6",
			cursorY:      12,
			mappedBuffer: []int{2, 4, 5, 7, 8, 12, 13},
			expected:     5,
		},
		{
			name:         "Test 7",
			cursorY:      5,
			mappedBuffer: []int{1, 2, 3, 4, 5, 6, 7},
			expected:     4,
		},
		{
			name:         "Test 8",
			cursorY:      6,
			mappedBuffer: []int{1, 2, 3, 4, 5, 6, 7},
			expected:     5,
		},
		{
			name:         "Test 9",
			cursorY:      7,
			mappedBuffer: []int{2, 4, 5, 7, 8, 13, 14},
			expected:     3,
		},
		{
			name:         "Test 10",
			cursorY:      10,
			mappedBuffer: []int{2, 4, 5, 7, 8, 13, 14},
			expected:     5,
		},
		{
			name:         "Test 11",
			cursorY:      13,
			mappedBuffer: []int{2, 4, 5, 7, 8, 13, 14},
			expected:     5,
		},
		{
			name:         "Test 12",
			cursorY:      1,
			mappedBuffer: []int{2, 4, 5, 7, 8, 13, 14},
			expected:     0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ret := fileeditor.CalcBufferLineFromACY(tc.cursorY, tc.mappedBuffer)

			if ret != tc.expected {
				t.Fatalf("Expected: %d, got: %d\n", tc.expected, ret)
			}
		})
	}
}

func TestCalcBufferIndexFromACXY(t *testing.T) {
	tests := []struct {
		name         string
		acY          int
		acX          int
		bufferLine   int
		visualBuffer []string
		mappedBuffer []int
		expected     int
	}{
		{
			name:       "Test 1",
			acY:        1,
			acX:        74,
			bufferLine: 0,
			visualBuffer: []string{
				"Hello, this is my text file that I want to edit! Something bout you makes",
				"me feel like a dangerous woman, something bout sometbout some you",
				"",
				"And Iknow this isnt wanhat oyu want but hey i dont give a fuck",
				"",
				"i hate this new lol 14.10 patch is so fking bs. Why are the adjusted adc i",
				"tems so fking trash. The components are horrendous and so awkward to build",
				" and it just makes the early game feel so weak for adcs!!!!!!",
				"",
			},
			mappedBuffer: []int{1, 2, 3, 4, 5, 8, 9},
			expected:     73,
		},
		{
			name:       "Test 2",
			acY:        8,
			acX:        44,
			bufferLine: 5,
			visualBuffer: []string{
				"Hello, this is my text file that I want to edit! Something bout you makes",
				"me feel like a dangerous woman, something bout sometbout some you",
				"",
				"And Iknow this isnt wanhat oyu want but hey i dont give a fuck",
				"",
				"i hate this new lol 14.10 patch is so fking bs. Why are the adjusted adc i",
				"tems so fking trash. The components are horrendous and so awkward to build",
				" and it just makes the early game feel so weak for adcs!!!!!!",
				"",
			},
			mappedBuffer: []int{1, 2, 3, 4, 5, 8, 9},
			expected:     191,
		},
		{
			name:       "Test 3",
			acY:        10,
			acX:        40,
			bufferLine: 5,
			visualBuffer: []string{
				"Hello, this is my text file that I want to edit! Something bo",
				"ut you makes",
				"me feel like a dangerous woman, something bout sometbout some",
				" you",
				"",
				"And Iknow this isnt wanhat oyu want but hey i dont give a fuc",
				"k",
				"",
				"i hate this new lol 14.10 patch is so fking bs. Why are the a",
				"djusted adc items so fking trash. The components are horrendo",
				"us and so awkward to build and it just makes the early game f",
				"eel so weak for adcs!!!!!!",
				"",
			},
			mappedBuffer: []int{2, 4, 5, 7, 8, 12, 13},
			expected:     100,
		},
		{
			name:       "Test 4",
			acY:        1,
			acX:        1,
			bufferLine: 0,
			visualBuffer: []string{
				"Hello, this is my text file that I want to edit! Something bout you makes",
				"me feel like a dangerous woman, something bout sometbout some you",
				"",
				"And Iknow this isnt wanhat oyu want but hey i dont give a fuck",
				"",
				"i hate this new lol 14.10 patch is so fking bs. Why are the adjusted adc i",
				"tems so fking trash. The components are horrendous and so awkward to build",
				" and it just makes the early game feel so weak for adcs!!!!!!",
				"",
			},
			mappedBuffer: []int{1, 2, 3, 4, 5, 8, 9},
			expected:     0,
		},
		{
			name:       "Test 5",
			acY:        12,
			acX:        9,
			bufferLine: 5,
			visualBuffer: []string{
				"Hello, this is my text file that I want to edit! Something bo",
				"ut you makes",
				"me feel like a dangerous woman, something bout sometbout some",
				" you",
				"",
				"And Iknow this isnt wanhat oyu want but hey i dont give a fuc",
				"k",
				"",
				"i hate this new lol 14.10 patch is so fking bs. Why are the a",
				"djusted adc items so fking trash. The components are horrendo",
				"us and so awkward to build and it just makes the early game f",
				"eel so weak for adcs!!!!!!",
				"",
			},
			mappedBuffer: []int{2, 4, 5, 7, 8, 12, 13},
			expected:     191,
		},
		{
			name:       "Test 6",
			acY:        4,
			acX:        1,
			bufferLine: 2,
			visualBuffer: []string{
				"Hello, this is my text file that I want to edit! Something bo",
				"ut you makes",
				"   like a dangerous woman, something bout sometbout some",
				"",
				"fancy, uuu, oooh, fancy u",
				"",
				"And Iknow this isnt wanhat oyu want but hey i dont give a fuc",
				"k",
				"",
				"i hate this new lol 14.10 patch is so fking bs. Why are the a",
				"djusted adc items so fking trash. The components are horrendo",
				"us and so awkward to build and it just makes the early game f",
				"eel so weak for adcs!!!!!!",
				"",
			},
			mappedBuffer: []int{2, 3, 4, 5, 6, 7, 8, 9, 13},
			expected:     0,
		},
		{
			name:       "Test 7",
			acY:        5,
			acX:        37,
			bufferLine: 3,
			visualBuffer: []string{
				"Hello, this is my text file that I want to edit! Something bo",
				"ut you makes",
				"   like a dangerous woman, something bout sometbout some",
				"",
				"fancy, uuu, oooh, fancy u",
				"",
				"And Iknow this isnt wanhat oyu want but hey i dont give a fuc",
				"k",
				"",
				"i hate this new lol 14.10 patch is so fking bs. Why are the a",
				"djusted adc items so fking trash. The components are horrendo",
				"us and so awkward to build and it just makes the early game f",
				"eel so weak for adcs!!!!!!",
				"",
			},
			mappedBuffer: []int{2, 3, 4, 5, 6, 7, 8, 9, 13},
			expected:     36,
		},
		{
			name:       "Test 8",
			acY:        1,
			acX:        1,
			bufferLine: 0,
			visualBuffer: []string{
				"",
			},
			mappedBuffer: []int{1},
			expected:     0,
		},
		{
			name:       "Test 9",
			acY:        2,
			acX:        2,
			bufferLine: 1,
			visualBuffer: []string{
				"",
				"",
				"",
			},
			mappedBuffer: []int{1, 3},
			expected:     1,
		},
		{
			name:       "Test 10",
			acY:        3,
			acX:        1,
			bufferLine: 1,
			visualBuffer: []string{
				"",
				"Something something balh alah nilah is such",
				" a good adc",
				"FEAST OR FAMINE!!! It's a party in the USA!",
			},
			mappedBuffer: []int{1, 3, 4},
			expected:     43,
		},
		{
			name:       "Test 11",
			acY:        1,
			acX:        1,
			bufferLine: 0,
			visualBuffer: []string{
				"Hello, this is my text file that I want to edit! Something bout you makes",
				"me feel like a dangerous woman, something bout sometbout some you",
				"",
				"And Iknow this isnt wanhat oyu want but hey i dont give a fuck",
				"",
				"i hate this new lol 14.10 patch is so fking bs. Why are the adjusted adc i",
				"tems so fking trash. The components are horrendous and so awkward to build",
				" and it just makes the early game feel so weak for adcs!!!!!!",
				"",
			},
			mappedBuffer: []int{1, 2, 3, 4, 5, 8, 9},
			expected:     0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ret := fileeditor.CalcBufferIndexFromACXY(tc.acX, tc.acY, tc.bufferLine, tc.visualBuffer, tc.mappedBuffer)

			if ret != tc.expected {
				t.Fatalf("Expected: %d, got: %d\n", tc.expected, ret)
			}
		})
	}
}

func TestCalcNewACXY(t *testing.T) {
	tests := []struct {
		name           string
		newBufMapped   []int
		bufferLine     int
		bufferIndex    int
		newEditorWidth int
		expectedX      int
		expectedY      int
	}{
		{
			name:           "Test 1",
			newBufMapped:   []int{2, 4, 5, 7, 8, 12, 13},
			bufferLine:     5,
			bufferIndex:    191,
			newEditorWidth: 62,
			expectedX:      9,
			expectedY:      12,
		},
		{
			name:           "Test 2",
			newBufMapped:   []int{2, 4, 5, 7, 8, 12, 13},
			bufferLine:     0,
			bufferIndex:    73,
			newEditorWidth: 62,
			expectedX:      13,
			expectedY:      2,
		},
		{
			name:           "Test 3",
			newBufMapped:   []int{1, 2, 3, 4, 5, 8, 9},
			bufferLine:     5,
			bufferIndex:    100,
			newEditorWidth: 75,
			expectedX:      27,
			expectedY:      7,
		},
		{
			name:           "Test 4",
			newBufMapped:   []int{2, 4, 5, 7, 8, 12, 13},
			bufferLine:     5,
			bufferIndex:    100,
			newEditorWidth: 62,
			expectedX:      40,
			expectedY:      10,
		},
		{
			name:           "Test 5",
			newBufMapped:   []int{2, 4, 5, 6, 7, 11, 12},
			bufferLine:     5,
			bufferIndex:    191,
			newEditorWidth: 64,
			expectedX:      3,
			expectedY:      11,
		},
		{
			name:           "Test 6",
			newBufMapped:   []int{2, 4, 5, 7, 8, 12, 13},
			bufferLine:     0,
			bufferIndex:    0,
			newEditorWidth: 62,
			expectedX:      1,
			expectedY:      1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			x, y := fileeditor.CalcNewACXY(
				tc.newBufMapped,
				tc.bufferLine, tc.bufferIndex,
				tc.newEditorWidth,
			)

			if x != tc.expectedX {
				t.Fatalf("Expected x: %d, got: %d\n", tc.expectedX, x)
			}

			if y != tc.expectedY {
				t.Fatalf("Expected y: %d, got: %d\n", tc.expectedY, y)

			}
		})
	}
}
