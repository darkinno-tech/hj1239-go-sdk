package model

import (
	"encoding/binary"
	"fmt"
)

// FileUploadNotificationCode 文件上传通知
type FileUploadNotificationCode struct {
	FileNameLen uint8      `hj1239:"offset:0,len:1,desc:文件名长度"`
	FileName    []byte     `hj1239:"offset:1,len:var,desc:文件名"`
	FileSize    uint32     `hj1239:"offset:var,len:4,desc:文件大小"`
	FileMD5     [16]byte   `hj1239:"offset:var,len:16,desc:文件MD5"`
	TotalBlocks uint16     `hj1239:"offset:var,len:2,desc:总块数"`
}

func (f *FileUploadNotificationCode) Encode() ([]byte, error) {
	nameLen := len(f.FileName)
	if nameLen > 255 {
		nameLen = 255
	}
	totalSize := 1 + nameLen + 4 + 16 + 2
	buf := make([]byte, totalSize)
	buf[0] = uint8(nameLen)
	copy(buf[1:1+nameLen], f.FileName)
	binary.BigEndian.PutUint32(buf[1+nameLen:5+nameLen], f.FileSize)
	copy(buf[5+nameLen:21+nameLen], f.FileMD5[:])
	binary.BigEndian.PutUint16(buf[21+nameLen:23+nameLen], f.TotalBlocks)
	return buf, nil
}

func (f *FileUploadNotificationCode) Decode(b []byte) error {
	if len(b) < 1+4+16+2 {
		return fmt.Errorf("file upload notification: data too short: %d", len(b))
	}
	nameLen := int(b[0])
	minLen := 1 + nameLen + 4 + 16 + 2
	if len(b) < minLen {
		return fmt.Errorf("file upload notification: data too short for name length %d", nameLen)
	}
	f.FileName = make([]byte, nameLen)
	copy(f.FileName, b[1:1+nameLen])
	f.FileSize = binary.BigEndian.Uint32(b[1+nameLen : 5+nameLen])
	copy(f.FileMD5[:], b[5+nameLen:21+nameLen])
	f.TotalBlocks = binary.BigEndian.Uint16(b[21+nameLen : 23+nameLen])
	return nil
}

func (f *FileUploadNotificationCode) Size() int {
	return 1 + len(f.FileName) + 4 + 16 + 2
}

// FileDataBlockCode 文件数据块
type FileDataBlockCode struct {
	BlockIndex uint16 `hj1239:"offset:0,len:2,desc:块索引"`
	BlockLen   uint16 `hj1239:"offset:2,len:2,desc:块数据长度"`
	BlockData  []byte `hj1239:"offset:4,len:var,desc:块数据"`
}

func (f *FileDataBlockCode) Encode() ([]byte, error) {
	totalSize := 2 + 2 + len(f.BlockData)
	buf := make([]byte, totalSize)
	binary.BigEndian.PutUint16(buf[0:2], f.BlockIndex)
	binary.BigEndian.PutUint16(buf[2:4], f.BlockLen)
	copy(buf[4:], f.BlockData)
	return buf, nil
}

func (f *FileDataBlockCode) Decode(b []byte) error {
	if len(b) < 4 {
		return fmt.Errorf("file data block: data too short: %d", len(b))
	}
	f.BlockIndex = binary.BigEndian.Uint16(b[0:2])
	f.BlockLen = binary.BigEndian.Uint16(b[2:4])
	if 4+int(f.BlockLen) > len(b) {
		return fmt.Errorf("file data block: data overflow")
	}
	f.BlockData = make([]byte, f.BlockLen)
	copy(f.BlockData, b[4:4+f.BlockLen])
	return nil
}

func (f *FileDataBlockCode) Size() int { return 2 + 2 + int(f.BlockLen) }

// FileUploadCompleteCode 文件上传完成
type FileUploadCompleteCode struct {
	Result    uint8  `hj1239:"offset:0,len:1,desc:上传结果 0x01=成功 0x02=失败"`
	BlockMask []byte `hj1239:"offset:1,len:var,desc:缺失块位掩码"`
}

func (f *FileUploadCompleteCode) Encode() ([]byte, error) {
	totalSize := 1 + len(f.BlockMask)
	buf := make([]byte, totalSize)
	buf[0] = f.Result
	copy(buf[1:], f.BlockMask)
	return buf, nil
}

func (f *FileUploadCompleteCode) Decode(b []byte) error {
	if len(b) < 1 {
		return fmt.Errorf("file upload complete: data too short: %d", len(b))
	}
	f.Result = b[0]
	f.BlockMask = make([]byte, len(b)-1)
	copy(f.BlockMask, b[1:])
	return nil
}

func (f *FileUploadCompleteCode) Size() int { return 1 + len(f.BlockMask) }
