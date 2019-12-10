package main
 
import(
	"net"
	"os"
	"fmt"
	"strconv"
	"flag"
	"github.com/golang/protobuf/proto"
	pb"server/protoc"
)
/*IpNode为worker信息，并以链表的形式连接*/
type IpNode struct {
	ipstring string
	prostring string
	next *IpNode
	prev *IpNode
}
/*IpNode初始化*/
func CreateIpNode(ipstring string,proto string) *IpNode{
	return &IpNode{
		ipstring,
		proto,
		nil,
		nil,
	}
}
/*添加新worker时，首先要遍历链表，检查worker是否已存在*/
func AddIpNode(cur *IpNode,ip string,protocol string){
	add := cur.next
	for add.ipstring!=ip&&add!=cur{
		add=add.next
	}
	if add.ipstring==ip{
		return
	}
	new := CreateIpNode(ip,protocol)
	new.next = cur.next
	new.next.prev=new
	cur.next = new
	new.prev = cur
	fmt.Println(cur.next.ipstring," has registered")
}
/*删除worker前操作，通过server端与worker通信确认worker下线，并之后采取删除操作*/
func DetectWorker(protocol string,ip string) bool{
	if ip=="start"{
		fmt.Println("no worker")
		return true
	}
	conn,err := net.Dial(protocol,ip)
	if err != nil {
		return false
	}
	buffer :=make([]byte,1024)
	/*发送一次运算，用以确认worker是否正常操作*/
	var rawdata pb.ComData
	rawdata.Firstnum=strconv.Itoa(4396)
	rawdata.Secondnum=strconv.Itoa(2200)
	rawdata.Flag=pb.DataFlag_DATAINT32
	rawdata.Opr=pb.Opr_ADD

	bytedata,_ :=proto.Marshal(&rawdata)
	_, err = conn.Write(bytedata)
	if err != nil {
		return false
	}

	_,err = conn.Read(buffer)
	if err != nil {
		return false
	}
	return true
}
/*删除worker操作，删除前需要遍历，找到worker位置。并检查worker是否为当前分配节点cur，如果是需要将cur更换*/
func DelIpNode(cur *IpNode,ip string){
	var del *IpNode
	if cur.ipstring==ip{
		del=cur
	}else{
		del = cur.next
		for del.ipstring!=ip&&del!=cur{
			del=del.next
		}
		if del.ipstring!=ip{
			return
		}
	}

	workeron := DetectWorker(del.prostring,del.ipstring)
	if workeron==true{
		return
	}

	if cur==del{
		cur=cur.prev
	}
	del.prev.next=del.next
	del.next.prev=del.prev
	fmt.Println(del.ipstring,"has deleted")
}
/*
开始注册操作，此处不适用goroutine
1、链表信息读写不能同时操作
2、server负载极小，正常情况下仅需要记录信息，初始验证都由client操作，由server验证。
*/
func StartRegister(protocol string,serverip string){
	listen,err := net.Listen(protocol,serverip)
	if err != nil {
		fmt.Println("出现错误：",err.Error())
		os.Exit(1)
	}
	defer listen.Close()

	fmt.Println("server start working")

	for{
		conn,err := listen.Accept()
		if err!=nil{
			continue
		}
		buffer := make([]byte,1024)
		n,err := conn.Read(buffer)
		if err != nil {
			fmt.Println(conn.RemoteAddr().String(),"出现错误：",err)
			continue
		}
		var rawmsgtoserver pb.MsgToServer
		err = proto.Unmarshal(buffer[:n],&rawmsgtoserver)
		if err != nil {
			fmt.Println(conn.RemoteAddr().String(),"protobuf err=", err)
			continue	
		}
		/*对应client发送worker下线情况*/
		if rawmsgtoserver.Flag==pb.ServerFlag_REGETWORKER{
			DelIpNode(cur,rawmsgtoserver.Ipstring)
		}
		/*对应worker注册情况*/
		if rawmsgtoserver.Flag==pb.ServerFlag_REGISTER{
			AddIpNode(cur,rawmsgtoserver.Ipstring,rawmsgtoserver.Prostring)
		}else {
			/*client中get和reget统一到一种情况执行*/
			var rawmsgtoclient pb.MsgToClient
			/*对应当前无worker情况*/
			if cur.ipstring=="start" && cur.next.ipstring=="start"{
				rawmsgtoclient.Ipstring="start"
			}else{
				if cur.next.ipstring=="start"{
					cur=cur.next
				}
				rawmsgtoclient.Ipstring=cur.next.ipstring
				rawmsgtoclient.Prostring=cur.next.prostring
				if rawmsgtoclient.Ipstring!="start"{
					fmt.Println(conn.RemoteAddr().String(),"was assigned：",cur.next.ipstring)
				}
			}
			cur=cur.next
			bytemsgtoclient,err :=proto.Marshal(&rawmsgtoclient)
			if err != nil {
				fmt.Println("protobuf err=", err)
				continue	
			}
			_,err=conn.Write(bytemsgtoclient)
			if err != nil {
				fmt.Println("conn.Write err=", err)	
				continue
			}
		}
	
		conn.Close()
	}
}

var startnode *IpNode//初始节点，之后初始化
var cur *IpNode//当前分配节点

func main(){
	startnode = CreateIpNode("start","start")//初始节点记为start，用以统一无worker和存在worker情况
	startnode.next = startnode
	startnode.prev = startnode
	cur = startnode//当前节点开始为start节点

	var serverip string
	var protocol string
	flag.StringVar(&serverip, "i", "172.17.0.15:8888", "服务器地址，默认172.17.0.15:8888")
	flag.StringVar(&protocol, "p", "tcp", "协议，默认为tcp")
	flag.Parse()

	StartRegister(protocol,serverip)
}
