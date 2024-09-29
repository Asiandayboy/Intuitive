package fileeditor

import "github.com/Asiandayboy/CLITextEditor/util/math"

/*
Returns the soft-wrapped corresponding index (0-indexed) in the FileBuffer array from the
apparent cursor's Y position using the mappedBuffer.
Since the mappedBuffer is always sorted, a binary search is the perfect solution

- acX and acY must ignore margin spaces

- acX and acY are 1-indexed values
*/
func CalcBufferLineFromACY(acY int, mappedBuffer []int, viewportOffsetY int) int {
	var target int = math.Clamp(1, acY, acY+1) + viewportOffsetY

	if target < len(mappedBuffer) && mappedBuffer[target-1] == target { // O(1) check
		return target - 1
	}

	var left int = 0
	var right int = len(mappedBuffer) - 1
	var res int = 0

	for left <= right {
		mid := (left + right) / 2

		v := mappedBuffer[mid]

		if target == v {
			return mid
		}

		if target <= mappedBuffer[left] {
			return left
		}

		if target >= mappedBuffer[right] {
			return right
		}

		if right-left == 1 && target > mappedBuffer[left] && target <= mappedBuffer[right] {
			return right
		}

		if target < v {
			right = mid
		} else if target > v {
			left = mid
		}

	}

	return res
}

/*
Returns the soft-wrapped corresponding index (0-indexed) in a particular line in the FileBuffer
array from the apparent cursor position using the visualBuffer and mappedBuffer

- acX and acY must ignore margin spaces

- acX and acY are 1-indexed values
*/
func CalcBufferIndexFromACXY(
	acX, acY, bufferLine int,
	visualBuffer []string, mappedBuffer []int,
	viewportOffsetY int,
) int {
	var totalLength int = 0

	var start int = 0
	var end int = mappedBuffer[bufferLine]

	if bufferLine > 0 {
		start = mappedBuffer[bufferLine-1]
	}

	/*
		Adds up each full-lengthed lines from the starting line, up to
		the buffer line, in which case acX is added instead and returns
		totalLength, which is the index for that line
	*/
	for i := start; i < end; i++ {
		if i == acY-1+viewportOffsetY {
			totalLength += acX
			break
		}

		totalLength += len(visualBuffer[i])
	}

	return totalLength - 1
}

/*
Returns the new soft-wrapped apparent cursor position after a window resize has occurred. The new X and Y
values will allow us to dynamically position the cursor so that it stays in the same spot in
the VisualBuffer instead of fixed on the screen in its previous position. Therefore, when the
window resizes, the cursor will move accordingly to match the resize.

The returned acX and acY values are 1-indexed, and do not take into account the viewport offsets

- bufferIndex and bufferLine are 0-indexed values

*/
func CalcNewACXY(
	newBufMapped []int, bufferLine int,
	bufferIndex int, newEditorWidth int,
	viewportOffsetY int,
) (newACX int, newACY int) {

	var start int = 0
	var end int = newBufMapped[bufferLine]

	if bufferLine > 0 {
		start = newBufMapped[bufferLine-1]
	}

	bufferIndex += 1

	var y int = 0
	var max int = bufferIndex
	for i := start; i < end; i++ {
		if max < newEditorWidth {
			y++
			break
		}
		max -= newEditorWidth
		y++
	}

	newACY = start + y - viewportOffsetY
	newACX = bufferIndex % (newEditorWidth - 1)

	if newACX == 0 {
		newACX = newEditorWidth - 1
	}

	return newACX, newACY
}

/*
	This func updates the buffer indicies based on the cursor's apparent position and
	the viewport offset using the visual buffer and the visual buffer mapped.
*/
func (f *FileEditor) UpdateBufferIndicies() {
	if f.SoftWrapEnabled {
		f.bufferLine = CalcBufferLineFromACY(f.apparentCursorY, f.VisualBufferMapped, f.ViewportOffsetY)
		f.bufferIndex = CalcBufferIndexFromACXY(
			f.apparentCursorX-EditorLeftMargin+1, f.apparentCursorY,
			f.bufferLine, f.VisualBuffer, f.VisualBufferMapped, f.ViewportOffsetY,
		)
	} else {
		f.bufferLine = f.apparentCursorY + f.ViewportOffsetY - 1
		f.bufferIndex = f.apparentCursorX + f.ViewportOffsetX - EditorLeftMargin
	}
}

/*
	This func align the buffer index to match the length of the FileBuffer instead of the visual
	buffer, which is inflated because of the fact tabs are replaced with spaces, which counts as an
	individual character and therefore contributes to the increase of the length of each line in
	the visual buffer corresponding to the line the FileBuffer
*/
func AlignBufferIndex(bufferIndex int, bufferLine int, tabMap TabMapType) (actualBufferIndex int) {
	indicies, exists := tabMap[bufferLine]
	if !exists {
		return bufferIndex
	}

	actualBufferIndex = bufferIndex
	for _, tabInfo := range indicies {
		end := tabInfo.End
		start := tabInfo.Start
		dif := end - start

		if end > bufferIndex {
			break
		}

		actualBufferIndex -= dif
	}

	return actualBufferIndex
}

func (f *FileEditor) IncrementCursorY() {
	f.apparentCursorY = math.Clamp(f.apparentCursorY+1, 1, f.GetViewportHeight())
}

func (f *FileEditor) DecrementCursorY() {
	f.apparentCursorY = math.Clamp(f.apparentCursorY-1, 1, f.GetViewportHeight())
}

/*
This function calculates the new cursor position for when soft wrap
is enabled. It is called every time a window resize occurs or whenever
soft wrap is toggled to true
*/
func (f *FileEditor) CalculateNewCursorPos() {
	newACX, newACY := CalcNewACXY(
		f.VisualBufferMapped, f.bufferLine,
		f.bufferIndex, f.GetViewportWidth(), f.ViewportOffsetY,
	)

	indexCheck := CalcBufferIndexFromACXY(
		newACX, newACY,
		f.bufferLine, f.VisualBuffer, f.VisualBufferMapped,
		f.ViewportOffsetY,
	)

	if indexCheck != f.bufferIndex {
		newACX = f.GetViewportWidth()
	}

	f.apparentCursorX = newACX + EditorLeftMargin - 1
	f.apparentCursorY = newACY
}
