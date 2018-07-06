package main

import (
	"net/http"
	"os"
	"io"
	"fmt"
	"html/template"
	"time"
	"encoding/binary"
)

const (
	DataPath ="D://picStore/data.txt"
	indexPath="D://picStore/index.txt"
	)

type fileIndex struct {
	name string
	offsite uint64
	size uint64
}

func(self *fileIndex) saveTofile(findex *os.File)  {
	//写文件时间戳
	findex.Write([]byte(self.name))
	//文件偏移量
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(self.offsite))
	findex.Write(buf)
	//文件大小
	var buf2 = make([]byte, 8)
	binary.BigEndian.PutUint64(buf2, uint64(self.size))
	findex.Write(buf2)

}
func(self *fileIndex) readFileIndex(findex *os.File,i int64) *fileIndex {
	name:=make([]byte,14)
	findex.ReadAt(name,i)
	picname:=string(name)
	var buf = make([]byte, 8)
	findex.ReadAt(buf,i+14)
	offindex:=uint64(binary.BigEndian.Uint64(buf))
	findex.ReadAt(buf,i+22)
	size:=uint64(binary.BigEndian.Uint64(buf))
	self.name=picname
	self.size=size
	self.offsite=offindex
	return  self
}

func String()string  {
	return ""
}

func main() {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		t,e:=template.ParseFiles("./fileupload.html")
		if e!=nil{
			panic(e)
		}else {
			t.Execute(writer,nil)
		}
	})

	//上传
	http.HandleFunc("/upload", func(writer http.ResponseWriter, request *http.Request) {
			err:=request.ParseMultipartForm(32<<20)
			if err==nil{
				file,_,_:=request.FormFile("pic")
				f,er:=os.OpenFile(DataPath,os.O_APPEND,0755)
				if er==nil{
					size,_:=io.Copy(f,file)
					t:=time.Now().Format("20060102150405")
					fmt.Println("文件名称",t)
					fmt.Println("文件大小",size)
					findex,_:=os.OpenFile(indexPath,os.O_APPEND,0755)
					n,_:=findex.Seek(0,2)
					fmt.Println("index文件",n)
					if n>0{
						//这里写入的顺序是filename offsite  size
						//计算新的index等于上一个文件的offsite+size
						var off = make([]byte, 8)
						findex.ReadAt(off,n-16)
						o:=binary.BigEndian.Uint64(off)
						var lastsize = make([]byte, 8)
						findex.ReadAt(lastsize,n-8)
						s:=binary.BigEndian.Uint64(lastsize)
						nindex:=o+s
						fmt.Println("新的偏移量是",nindex)
						//写文件时间戳
						findex.Write([]byte(t))
						var buf = make([]byte, 8)
						binary.BigEndian.PutUint64(buf, uint64(nindex))
						findex.Write(buf)
						//文件大小
						var buf2 = make([]byte, 8)
						binary.BigEndian.PutUint64(buf2, uint64(size))
						findex.Write(buf2)
					}else{
						
						//写文件时间戳
						findex.Write([]byte(t))
						//文件偏移量
						var buf = make([]byte, 8)
						binary.BigEndian.PutUint64(buf, uint64(0))
						fmt.Println("off",buf)
						findex.Write(buf)
						//文件大小
						var buf2 = make([]byte, 8)
						binary.BigEndian.PutUint64(buf2, uint64(size))
						fmt.Println(buf2)
						findex.Write(buf2)
					}
					defer findex.Close()
				}
				defer  f.Close()
			}else{
				panic("not multipart file")
			}
	})
	//获取
	http.HandleFunc("/getpic", func(writer http.ResponseWriter, request *http.Request) {
		pic:=request.FormValue("pic")
		fmt.Println("查找的图片是",pic)
		findex,_:=os.OpenFile(indexPath,os.O_APPEND|os.O_RDWR,0755)
		//获取文件总长度
		n,_:=findex.Seek(0,os.SEEK_END)
		for i:=int64(0); i<n;i+=30{
			ni,_:=findex.Seek(i,0)
			fmt.Println("当前查找的是",ni,"i是",i)
			name:=make([]byte,14)
			findex.ReadAt(name,ni)
			picname:=string(name)
			fmt.Println("图片名称是",picname)
			if picname==pic{
				var buf = make([]byte, 8)
				findex.ReadAt(buf,i+14)
				offindex:=uint64(binary.BigEndian.Uint64(buf))
				fmt.Println("偏移量是",offindex)
				findex.ReadAt(buf,i+22)
				size:=uint64(binary.BigEndian.Uint64(buf))
				fmt.Println("大小是",size)
				f,_:=os.OpenFile(DataPath,os.O_RDONLY,0755)
				f.Seek(int64(offindex),0)
				var img=make([]byte,size)
				f.ReadAt(img,int64(offindex))
				//io.Copy(writer,f)
				writer.Write(img)
				defer f.Close()
				break
			}
		}

		defer  findex.Close()
	})
	http.ListenAndServe(":8888",nil)

}
