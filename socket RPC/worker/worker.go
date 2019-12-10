package main
 
import(
	"net"
	"github.com/golang/protobuf/proto"
	"fmt"
	"flag"
	"strconv"
	pb"worker/protoc"
)
/*注册模块，负责去server端注册，注册成功后就断开连接*/
func register()bool{
	conn,err := net.Dial(protocol,serverip)
	if err != nil {
		fmt.Println("register failed\n", err)
		return false
	}
	var rawmsgtoserver pb.MsgToServer
	/*发送本地外网地址及端口*/
	rawmsgtoserver.Ipstring = extranetip
	rawmsgtoserver.Prostring = protocol
	rawmsgtoserver.Flag = pb.ServerFlag_REGISTER

	bytemsgtoserver,err :=proto.Marshal(&rawmsgtoserver)
	if err != nil {
		fmt.Println("register failed\n", err)
		return false
	}
	_, err = conn.Write(bytemsgtoserver)
	if err != nil {
		fmt.Println("register failed\n", err)
		return false
	}
	conn.Close()
	return true
}
 
/*主要进程，负责与client交互*/
func handleconn(conn net.Conn){
	for{
		buffer := make([]byte,1024)
		n,err := conn.Read(buffer)
		if err != nil {
			fmt.Println(conn.RemoteAddr().String()," done")
			return
		}
		var rawdata pb.ComData
		err = proto.Unmarshal(buffer[:n],&rawdata)
		var rawres pb.ComRes
		/*负责将string转化为数字进行运算操作，并将结果转化为string发送给client*/
		if rawdata.Flag==pb.DataFlag_DATAINT32{
			firstnum,_ := strconv.ParseInt(rawdata.Firstnum,10,32)
			secondnum,_ := strconv.ParseInt(rawdata.Secondnum,10,32)
			if secondnum==0&&rawdata.Opr==pb.Opr_DIV{
				rawres.Flag=pb.ResFlag_FALSE
				rawres.Res=strconv.FormatInt(0,10)
			}else{
				rawres.Flag=pb.ResFlag_RESINT64
				var res int64
				if rawdata.Opr==pb.Opr_ADD{
					res=firstnum+secondnum
				}else if rawdata.Opr==pb.Opr_SUB{
					res=firstnum-secondnum
				}else if rawdata.Opr==pb.Opr_MUL{
					res=firstnum*secondnum
				}else {
					res=firstnum/secondnum
				}
				rawres.Res=strconv.FormatInt(res,10)
			}
		}else if rawdata.Flag==pb.DataFlag_DATAINT64{
			firstnum,_ := strconv.ParseInt(rawdata.Firstnum,10,64)
			secondnum,_ := strconv.ParseInt(rawdata.Secondnum,10,64)
			if secondnum==0&&rawdata.Opr==pb.Opr_DIV{
				rawres.Flag=pb.ResFlag_FALSE
				rawres.Res=strconv.FormatInt(0,10)
			}else{
				rawres.Flag=pb.ResFlag_RESINT64
				var res int64
				if rawdata.Opr==pb.Opr_ADD{
					res=firstnum+secondnum
				}else if rawdata.Opr==pb.Opr_SUB{
					res=firstnum-secondnum
				}else if rawdata.Opr==pb.Opr_MUL{
					res=firstnum*secondnum
				}else {
					res=firstnum/secondnum
				}
				rawres.Res=strconv.FormatInt(res,10)
			}
		}else{
			firstnum,_ := strconv.ParseFloat(rawdata.Firstnum,64)
			secondnum,_ := strconv.ParseFloat(rawdata.Secondnum,64)
			if secondnum==0&&rawdata.Opr==pb.Opr_DIV{
				rawres.Flag=pb.ResFlag_FALSE
				rawres.Res=strconv.FormatFloat(0,'E',-1,64)
			}else{
				rawres.Flag=pb.ResFlag_RESFLOAT
				var res float64
				if rawdata.Opr==pb.Opr_ADD{
					res=firstnum+secondnum
				}else if rawdata.Opr==pb.Opr_SUB{
					res=firstnum-secondnum
				}else if rawdata.Opr==pb.Opr_MUL{
					res=firstnum*secondnum
				}else {
					res=firstnum/secondnum
				}
				rawres.Res=strconv.FormatFloat(res,'E',-1,64)
			}
			
		}
		byteres,err :=proto.Marshal(&rawres)
		if err != nil {
			fmt.Println("marshal msg to ",conn.RemoteAddr().String()," failed,maybe client kill it")
		}
		_, err = conn.Write(byteres)
		if err != nil {
			fmt.Println("send msg to ",conn.RemoteAddr().String()," failed,maybe client kill it")
		}
	}
}
 
var intranetip string//内网ip及端口，负责本地监听
var extranetip string//外网ip及端口，负责client连接
var protocol string//协议
var serverip string//服务器ip

func main() {

	flag.StringVar(&intranetip, "i", "172.17.0.15:9999", "内网口,默认为172.17.0.15:9999")
	flag.StringVar(&extranetip, "e", "122.51.83.192:9999", "外网口,默认为122.51.83.192:9999")
	flag.StringVar(&protocol, "p", "tcp", "协议，默认为tcp")
	flag.StringVar(&serverip, "s", "122.51.83.192:8888", "服务器地址，默认122.51.83.192:8888")
	flag.Parse()
	/*如果注册失败，则worker不能上线。*/
	done := register()
	if done==false{
		return
	}
	/*监听外网端口*/
	listen,_ := net.Listen(protocol,intranetip)
	defer listen.Close()
	fmt.Println(extranetip,"waiting client")
	for{
		conn,err := listen.Accept()
		if err != nil {
			continue
		}
		connip := conn.RemoteAddr().String()
		fmt.Println(connip,"connect to",extranetip)
		/*goroutine负责协程操作*/
		go handleconn(conn)
	}
}

