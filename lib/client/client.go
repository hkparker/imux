package client

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/hkparker/TLJ"
	"github.com/hkparker/imux/lib/common"
	"os"
	"reflect"
	"strings"
)

func CommandLoop(control tlj.Client, workers []tlj.StreamWriter, chunk_size int) {
	stdin := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("imux> ")
		line, _ := stdin.ReadString('\n')
		text := strings.TrimSpace(line)
		cmd := strings.Fields(text)
		if len(cmd) == 0 {
			continue
		}
		command := cmd[0]
		var args []string
		if len(command) > 1 {
			args = cmd[1:]
		}
		if command == "get" {
			//common.RequestFiles(args)
			// send a Command{} with get and the files as args (server wont respond (or does it need to respond when all done and with updates?), will just stream chunks down nonced workers)
			// file finished are send by write buffer to global current_transfer chan
			// speed updates are 1 per second?  need to ask every worker?  (workers update global speed store, sum that)
			//PrintProgress(file_finished, speed_update, all_done)
			// get updates from server?
		} else if command == "put" {
			file_list, total_bytes := common.ParseFileList(args)
			common.UploadFiles(file_list, total_bytes, workers, chunk_size)
		} else if command == "exit" {
			control.Request(common.Command{
				Command: "exit",
			})
			control.Dead <- errors.New("user exit")
			break
		} else {
			req, err := control.Request(common.Command{
				Command: command,
				Args:    args,
			})
			if err != nil {
				go func() {
					control.Dead <- errors.New(fmt.Sprintf("error sending command: %v", err))
				}()
				break
			}
			command_output := make(chan string)
			req.OnResponse(reflect.TypeOf(common.Message{}), func(iface interface{}) {
				if message, ok := iface.(*common.Message); ok {
					command_output <- message.String
				}
			})
			fmt.Println(<-command_output)
		}
	}
}
