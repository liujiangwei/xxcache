package redis

import (
"bytes"
"github.com/go-ini/ini"
"strings"
)

func ParseInfo(rawStr string) (redisInfo *Info, err error){
	file := ini.Empty()
	rawBuf := bytes.NewBuffer([]byte(rawStr))
	var lines []string

	var line []byte
	for {
		if line, err = rawBuf.ReadBytes('\n'); err != nil{
			break // io.EOF return
		}

		str := strings.TrimSpace(string(line))
		if str == ""{
			continue
		}

		if str[0] == '#'{
			line := bytes.TrimSpace(line[1:])
			lines = append(lines, "[" + strings.ToLower(string(line)) + "]")
			continue
		}

		lines = append(lines, strings.Replace(str, ":", "=", 1))
	}

	file , err = ini.Load([]byte(strings.Join(lines, "\n")))
	if err != nil{
		return nil, err
	}

	repl := Replication{}
	repl.parse(file.Section("replication"))

	return &Info{
		Replication:repl,
	}, nil
}

type Info struct {
	Replication Replication
}

type Replication struct {
	Role                       string
	ConnectedSlaves            int
	MasterReplicationId        string
	MasterReplicationId2       string
	MasterReplOffset           int
	SecondReplOffset           int
	replBacklogActive          int
	replBacklogSize            int
	replBacklogFirstByteOffset int
	replBacklogHistLen         int
}

func (repl *Replication) parse(section *ini.Section) {
	repl.Role =  section.Key("role").MustString("master")
	repl.ConnectedSlaves = section.Key("connected_slaves").MustInt(0)
	repl.MasterReplicationId =  section.Key("master_replid").MustString("")
	repl.MasterReplicationId2 = section.Key("master_replid2").MustString("")
	repl.MasterReplOffset = section.Key("master_repl_offset").MustInt(0)
	repl.SecondReplOffset = section.Key("second_repl_offset").MustInt(-1)
}