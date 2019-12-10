package main

import(
	"github.com/golang/protobuf/proto"
	"github.com/schollz/progressbar"
	"fmt"
	"net"
	"strconv"
	"flag"
	"os"
	pb"client/protoc"
	
)

/*第一次连接server，获取worker的ip、port、protocol*/
func getworker(){
	/*连接server，并返回socket口conn*/
	conn,err := net.Dial(serverpro,serverip)
	if err != nil {
		fmt.Println("connect server failed", err)
		return 
	}
	/*rawmsgtoserver为发送给server信息，通过flag位代表第一次连接，需要获取worker信息*/
	var rawmsgtoserver pb.MsgToServer
	rawmsgtoserver.Flag = pb.ServerFlag_GETWORKER
	bytemsgtoserver,err :=proto.Marshal(&rawmsgtoserver)
	if err != nil {
		fmt.Println("protobuf err=", err)
		return	
	}
	_,err=conn.Write(bytemsgtoserver)
	if err != nil {
		fmt.Println("conn.Write err=", err)	
		return
	}
	/*接收从server发送的信息，并确认worker信息*/
	buffer := make([]byte,1024)
	n,err := conn.Read(buffer)
	if err != nil {
		fmt.Println("connect server failed：",err)
		return
	}
	var rawmsgtoclient pb.MsgToServer
	err = proto.Unmarshal(buffer[:n],&rawmsgtoclient)
	if err != nil {
		fmt.Println("protobuf err=", err)
		return
	}

	workerip=rawmsgtoclient.Ipstring
	workerpro= rawmsgtoclient.Prostring
	conn.Close()
}
/*如果出现worker无法运行，则向server发送当前worker信息，并请求获取新worker*/
func regetworker(protocol string,ip string){
	conn,err := net.Dial(serverpro,serverip)
	if err != nil {
		fmt.Println("connect server failed", err)
		workerip="start"
		workerpro="start"
		return 
	}
	/*发送消息相比getworker，需要额外添加当前worker信息，并置flag位为reget*/
	var rawmsgtoserver pb.MsgToServer
	rawmsgtoserver.Flag = pb.ServerFlag_REGETWORKER
	rawmsgtoserver.Ipstring = ip
	rawmsgtoserver.Prostring=protocol
	bytemsgtoserver,err :=proto.Marshal(&rawmsgtoserver)
	if err != nil {
		fmt.Println("protobuf err=", err)
		return	
	}
	_,err=conn.Write(bytemsgtoserver)
	if err != nil {
		fmt.Println("conn.Write err=", err)	
		return
	}
	/*接受消息同getworker*/
	buffer := make([]byte,1024)
	n,err := conn.Read(buffer)
	if err != nil {
		fmt.Println("connect server failed：",err)
		return
	}
	var rawmsgtoclient pb.MsgToServer
	err = proto.Unmarshal(buffer[:n],&rawmsgtoclient)
	if err != nil {
		fmt.Println("protobuf err=", err)
		return
	}

	workerip=rawmsgtoclient.Ipstring
	workerpro= rawmsgtoclient.Prostring
	conn.Close()
}

/*主要功能函数，负责发送、接受及可视化等操作*/
func handleconn(protocol string,ip string){
	/*由于server在没有worker时，会将worker的ip回复成start*/
	if ip=="start"{
		fmt.Println("no worker")
		return
	}
	/*连接worker*/
	conn,err := net.Dial(protocol,ip)
	if err != nil {
		regetworker(protocol,ip)
		defer handleconn(workerpro,workerip)
		return 
	}
	buffer :=make([]byte,1024)
	/*percent负责可视化操作进度*/
	percent := count_all/100
	/*负责将运算结果输出到本地文件*/
	fd,err := os.OpenFile(filename,os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
	/*模拟用户百万级操作*/
	for ; count >0; count--{
		/*可视化操作，详情可查 github.com/schollz/progressbar */
		if printpro=="y" {
			if count%percent==0{
				bar.Add(1)
			}
		}
		/*首先将数字转化为string，统一传输。在worker端重新转化数字*/
		var rawdata pb.ComData
		if count%3==0 {
			rawdata.Firstnum=strconv.Itoa(4396)
			rawdata.Secondnum=strconv.Itoa(2200)
			rawdata.Flag=pb.DataFlag_DATAINT32
		}else if count%3==1 {
			rawdata.Firstnum=strconv.FormatInt(4396,10)
			rawdata.Secondnum=strconv.FormatInt(2200,10)
			rawdata.Flag=pb.DataFlag_DATAINT64
		}else{
			rawdata.Firstnum=strconv.FormatFloat(4396.4396,'E',-1,64)
			rawdata.Secondnum=strconv.FormatFloat(2200.2200,'E',-1,64)
			rawdata.Flag=pb.DataFlag_DATAFLOAT
		}
		/*运算结果写入本地文件*/
		_,err = fd.WriteString("first num: "+rawdata.Firstnum+" second num: "+rawdata.Secondnum)
		if err !=nil{
			fmt.Println(err)
			continue
		}
		/*负责加减乘除*/
		switch(count%4){
		case 0:
			rawdata.Opr=pb.Opr_ADD
			_,err = fd.WriteString(" Opr: ADD\n")
		case 1:
			rawdata.Opr=pb.Opr_SUB
			_,err = fd.WriteString(" Opr: SUB\n")
		case 2:
			rawdata.Opr=pb.Opr_MUL
			_,err = fd.WriteString(" Opr: MUL\n")
		default:
			rawdata.Opr=pb.Opr_DIV
			_,err = fd.WriteString(" Opr: DIV\n")
		}
		/*发送信息至worker，如果发送失败则通知server，请求更换worker*/
		if err !=nil{
			fmt.Println(err)
			continue
		}
		bytedata,err :=proto.Marshal(&rawdata)
		_, err = conn.Write(bytedata)
		if err != nil {
			regetworker(protocol,ip)
			fd.Close()
			defer handleconn(workerpro,workerip)
			return
		}
		/*从worker接受信息，如果发送失败则通知server，请求更换worker*/
		n,err := conn.Read(buffer)
		if err != nil {
			regetworker(protocol,ip)
			fd.Close()
			defer handleconn(workerpro,workerip)
			return
		}
		var rawres pb.ComRes
		err = proto.Unmarshal(buffer[:n],&rawres)
		switch(rawres.Flag){
		case pb.ResFlag_FALSE:
			_,err = fd.WriteString("false\n\n")
			if err !=nil{
				fmt.Println(err)
				continue
			}
		default:
			_,err = fd.WriteString(rawres.Res+"\n\n")
		}
		
	}
	if printpro=="true"{
		fmt.Println()
	}

}



var serverip string//server端ip及端口
var serverpro string//server端协议
var workerip string//worker端ip及端口
var workerpro string//worker端协议
var filename string//本地文件，负责记录运算过程及结果
var count int//当前运算次数
var count_all int//总共需要运算次数
var printpro string//是否需要打印进度（y/n）
var bar *progressbar.ProgressBar//进度条


func main() {
	flag.StringVar(&filename, "f", "1.txt", "log")
	flag.StringVar(&serverpro, "p", "tcp", "协议，默认为tcp")
	flag.StringVar(&serverip, "e", "122.51.83.192:8888", "服务器地址，默认122.51.83.192:8888")
	flag.StringVar(&printpro, "print", "n", "打印进度，y/n,默认n（不打印）")
	flag.Parse()
	/*初始化进度条*/
	bar = progressbar.New(100)
	count = 10000
	count_all=10000
	//regetworker("tcp","122.51.83.192:9999")
	getworker()
	handleconn(workerpro, workerip)
}
