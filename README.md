# picstore
use golang to store pic as txt file 
简单的建立了两个文件图片上传的时候一个文件保存实际的二进制图片数据，一个保存图片相关的索引数据(包括文件名称，在数据文件中的起始的偏移量，图片尺寸)
并且依据这些数据可以重新访问图片。
