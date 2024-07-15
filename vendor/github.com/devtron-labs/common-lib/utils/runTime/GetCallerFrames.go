package runTime

import (
	"runtime"
)

// GetCallerFileName - returns the file name of the invoked func
func GetCallerFileName() string {
	// Skip GetCurrentFunctionName
	return getFrame(1).File
}

// GetCallerLineNumber - returns the line number of the invoked func
func GetCallerLineNumber() int {
	// Skip GetCurrentFunctionName
	return getFrame(1).Line
}

// GetCallerFunctionName - returns the function name of the invoked func
func GetCallerFunctionName() string {
	// Skip GetCallerFunctionName and the function to get the caller of
	return getFrame(2).Function
}

// getFrame returns the runtime.Frame for the targetFrameIndex from the runtime caller stack
//
// examples:
//  1. getFrame(0) -> returns current method frame
//  2. getFrame(1) -> returns caller method frame
//  3. getFrame(2) -> returns caller's caller method frame
func getFrame(targetFrameIndex int) runtime.Frame {
	// Set size to targetFrameIndex + 2 to ensure we have room for function runtime.Callers and getFrame
	programCounters := make([]uintptr, targetFrameIndex+2)
	// We need the frame at index targetFrameIndex + 2, since we never want function runtime.Callers and getFrame
	n := runtime.Callers(2, programCounters)

	frame := runtime.Frame{Function: "unknown"}
	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])
		for more, frameIndex := true, 0; more && frameIndex <= targetFrameIndex; frameIndex++ {
			var frameCandidate runtime.Frame
			frameCandidate, more = frames.Next()
			if frameIndex == targetFrameIndex {
				frame = frameCandidate
			}
		}
	}

	return frame
}
