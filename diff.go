package git

/*
#include <git2.h>

extern int _go_git_diff_foreach(git_diff *diff, int eachFile, int eachHunk, int eachLine, void *payload);
*/
import "C"
import (
	"runtime"
	"unsafe"
)

const (
	DiffFlagBinary = C.GIT_DIFF_FLAG_BINARY
	DiffFlagNotBinary = C.GIT_DIFF_FLAG_NOT_BINARY
	DiffFlagValidOid = C.GIT_DIFF_FLAG_VALID_OID

	DeltaUnmodified = C.GIT_DELTA_UNMODIFIED
	DeltaAdded = C.GIT_DELTA_ADDED
	DeltaDeleted = C.GIT_DELTA_DELETED
	DeltaModified = C.GIT_DELTA_MODIFIED
	DeltaRenamed = C.GIT_DELTA_RENAMED
	DeltaCopied = C.GIT_DELTA_COPIED
	DeltaIgnored = C.GIT_DELTA_IGNORED
	DeltaUntracked = C.GIT_DELTA_UNTRACKED
	DeltaTypeChange = C.GIT_DELTA_TYPECHANGE

	DiffLineContext = C.GIT_DIFF_LINE_CONTEXT
	DiffLineAddition = C.GIT_DIFF_LINE_ADDITION
	DiffLineDeletion = C.GIT_DIFF_LINE_DELETION
	DiffLineContextEOFNL = C.GIT_DIFF_LINE_CONTEXT_EOFNL
	DiffLineAddEOFNL = C.GIT_DIFF_LINE_ADD_EOFNL
	DiffLineDelEOFNL = C.GIT_DIFF_LINE_DEL_EOFNL

	DiffLineFileHdr = C.GIT_DIFF_LINE_FILE_HDR
	DiffLineHunkHdr = C.GIT_DIFF_LINE_HUNK_HDR
	DiffLineBinary = C.GIT_DIFF_LINE_BINARY
)

type DiffFile struct {
	file C.git_diff_file
	Path string
}

func newDiffFile(file *C.git_diff_file) *DiffFile {
	return &DiffFile{
		file: *file,
		Path: C.GoString(file.path),
	}
}

func (df *DiffFile) Oid() *Oid {
	return newOidFromC(&df.file.oid)
}

func (df *DiffFile) Size() int {
	return int(df.file.size)
}

func (df *DiffFile) Flags() uint32 {
	return uint32(df.file.flags)
}

func (df *DiffFile) Mode() uint16 {
	return uint16(df.file.mode)
}

type DiffDelta struct {
	delta C.git_diff_delta
	OldFile *DiffFile
	NewFile *DiffFile
}

func newDiffDelta(delta *C.git_diff_delta) *DiffDelta {
	return &DiffDelta{
		delta: *delta,
		OldFile: newDiffFile(&delta.old_file),
		NewFile: newDiffFile(&delta.new_file),
	}
}

func (dd *DiffDelta) Status() int {
	return int(dd.delta.status)
}

func (dd *DiffDelta) Flags() uint32 {
	return uint32(dd.delta.flags)
}

func (dd *DiffDelta) Similarity() uint16 {
	return uint16(dd.delta.similarity)
}

type DiffHunk struct {
	hunk C.git_diff_hunk
	Header string
	DiffDelta
}

func newDiffHunk(delta *C.git_diff_delta, hunk *C.git_diff_hunk) *DiffHunk {
	return &DiffHunk{
		hunk: *hunk,
		Header: C.GoStringN(&hunk.header[0], C.int(hunk.header_len)),
		DiffDelta: *newDiffDelta(delta),
	}
}

func (dh *DiffHunk) OldStart() int {
	return int(dh.hunk.old_start)
}

func (dh *DiffHunk) OldLines() int {
	return int(dh.hunk.old_lines)
}

func (dh *DiffHunk) NewStart() int {
	return int(dh.hunk.new_start)
}

func (dh *DiffHunk) NewLines() int {
	return int(dh.hunk.new_lines)
}

type DiffLine struct {
	line C.git_diff_line
	Content string
	DiffHunk
}

func newDiffLine(delta *C.git_diff_delta, hunk *C.git_diff_hunk, line *C.git_diff_line) *DiffLine {
	return &DiffLine{
		line: *line,
		Content: C.GoStringN(line.content, C.int(line.content_len)),
		DiffHunk: *newDiffHunk(delta, hunk),
	}
}

func (dl *DiffLine) Origin() byte {
	return byte(dl.line.origin)
}

func (dl *DiffLine) OldLineno() int {
	return int(dl.line.old_lineno)
}

func (dl *DiffLine) NewLineno() int {
	return int(dl.line.new_lineno)
}

func (dl *DiffLine) NumLines() int {
	return int(dl.line.num_lines)
}

func (dl *DiffLine) ContentOffset() int {
	return int(dl.line.content_offset)
}

type Diff struct {
	ptr *C.git_diff
}

func newDiff(ptr *C.git_diff) *Diff {
	if ptr == nil {
		return nil
	}

	diff := &Diff{
		ptr: ptr,
	}

	runtime.SetFinalizer(diff, (*Diff).Free)
	return diff
}

func (diff *Diff) Free() {
	runtime.SetFinalizer(diff, nil)
	C.git_diff_free(diff.ptr)
}

func (diff *Diff) forEachFileWrap(ch chan *DiffDelta) {
	C._go_git_diff_foreach(diff.ptr, 1, 0, 0, unsafe.Pointer(&ch))
	close(ch)
}

func (diff *Diff) ForEachFile() chan *DiffDelta {
	ch := make(chan *DiffDelta, 0)
	go diff.forEachFileWrap(ch)
	return ch
}

//export diffForEachFileCb
func diffForEachFileCb(delta *C.git_diff_delta, progress C.float, payload unsafe.Pointer) int {
	ch := *(*chan *DiffDelta)(payload)

	select {
	case ch <-newDiffDelta(delta):
	case <-ch:
		return -1
	}

	return 0
}

func (diff *Diff) forEachHunkWrap(ch chan *DiffHunk) {
	C._go_git_diff_foreach(diff.ptr, 0, 1, 0, unsafe.Pointer(&ch))
	close(ch)
}

func (diff *Diff) ForEachHunk() chan *DiffHunk {
	ch := make(chan *DiffHunk, 0)
	go diff.forEachHunkWrap(ch)
	return ch
}

//export diffForEachHunkCb
func diffForEachHunkCb(delta *C.git_diff_delta, hunk *C.git_diff_hunk, payload unsafe.Pointer) int {
	ch := *(*chan *DiffHunk)(payload)

	select {
	case ch <-newDiffHunk(delta, hunk):
	case <-ch:
		return -1
	}

	return 0
}

func (diff *Diff) forEachLineWrap(ch chan *DiffLine) {
	C._go_git_diff_foreach(diff.ptr, 0, 0, 1, unsafe.Pointer(&ch))
	close(ch)
}

func (diff *Diff) ForEachLine() chan *DiffLine {
	ch := make(chan *DiffLine, 0)
	go diff.forEachLineWrap(ch)
	return ch
}

//export diffForEachLineCb
func diffForEachLineCb(delta *C.git_diff_delta, hunk *C.git_diff_hunk, line *C.git_diff_line, payload unsafe.Pointer) int {
	ch := *(*chan *DiffLine)(payload)

	select {
	case ch <-newDiffLine(delta, hunk, line):
	case <-ch:
		return -1
	}

	return 0
}

func (diff *Diff) NumDeltas() int {
	return int(C.git_diff_num_deltas(diff.ptr))
}

func (diff *Diff) GetDelta(index int) *DiffDelta {
	ptr := C.git_diff_get_delta(diff.ptr, C.size_t(index))
	if ptr == nil {
		return nil
	}

	return newDiffDelta(ptr)
}

func (diff *Diff) Patch(deltaIndex int) *Patch {
	var patchPtr *C.git_patch

	C.git_patch_from_diff(&patchPtr, diff.ptr, C.size_t(deltaIndex))

	return newPatch(patchPtr)
}