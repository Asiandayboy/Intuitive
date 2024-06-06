package fileeditor

/*
Returns the corresponding index (0-indexed) in the FileBuffer array from the
apparent cursor's Y position using the mappedBuffer.
Since the mappedBuffer is always sorted, a binary search is the perfect solution

- acX and acY must ignore margin spaces

- acX and acY are 1-indexed values
*/
func CalcBufferLineFromACY(acY int, mappedBuffer []int) int {
	var target int = acY

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
Returns the corresponding index (0-indexed) in a particular line in the FileBuffer
array from the apparent cursor position using the visualBuffer and mappedBuffer

- acX and acY must ignore margin spaces

- acX and acY are 1-indexed values
*/
func CalcBufferIndexFromACXY(acX, acY, bufferLine int, visualBuffer []string, mappedBuffer []int) int {
	var totalLength int = 0

	var start int = 0
	var end int = mappedBuffer[bufferLine]

	if bufferLine > 0 {
		start = mappedBuffer[bufferLine-1]
	}

	for i := start; i < end; i++ {
		if i == acY-1 {
			totalLength += acX
			break
		}

		totalLength += len(visualBuffer[i])
	}

	return totalLength - 1
}

/*
Returns the new apparent cursor position after a window resize has occurred. The new X and Y
values will allow us to dynamically position the cursor so that it stays in the same spot in
the VisualBuffer instead of fixed on the screen in its previous position. Therefore, when the
window resizes, the cursor will move accordingly to match the resize.

- bufferIndex and bufferLine are 0-indexed values

*/
func CalcNewACXY(
	newBufMapped []int, bufferLine int,
	bufferIndex int, newEditorWidth int,
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

	newACY = start + y
	newACX = bufferIndex % (newEditorWidth - 1)

	return newACX, newACY
}

func (e *FileEditor) UpdateBufferIndicies() {
	e.bufferLine = CalcBufferLineFromACY(e.apparentCursorY, e.VisualBufferMapped)
	e.bufferIndex = CalcBufferIndexFromACXY(
		e.apparentCursorX-EditorLeftMargin+1, e.apparentCursorY,
		e.bufferLine, e.VisualBuffer, e.VisualBufferMapped,
	)
}
