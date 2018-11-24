package cli_client

import (
	"context"
	"github.com/mame82/P4wnP1_go/common"
	"github.com/spf13/cobra"
	"fmt"
	"path/filepath"
	"strings"
	"io"
	"bufio"
	"os"
	"errors"
	"log"
	"strconv"
	pb "github.com/mame82/P4wnP1_go/proto"
)

var (
	tmpHidCommands = ""
	tmpRunStored   = ""
	tmpHidTimeout  = uint32(0) // values < 0 = endless
)

var hidCmd = &cobra.Command{
	Use:   "hid",
	Short: "Use keyboard or mouse functionality",
}

var hidRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a HID Script",
	Long:"Run script provided from standard input, commandline parameter or by path to script file on P4wnP1",
	Run: cobraHidRun,
}

var hidJobCmd = &cobra.Command{
	Use:   "job",
	Short: "Run a HID Script as background job",
	Long:"Run a background script provided from standard input, commandline parameter or by path to script file on P4wnP1",
	Run: cobraHidJob,
}

var hidJobCancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "Cancel background job given by its ID",
	Run: cobraHidJobCancel,
}

// Decision on how to run scripts (terms "local"/"remote" from perspective of gRPC server):
//
// 1) `P4wnP1_cli HID run` (no host flag, no additional args)
// Assume that the CLI runs on the same host as the server, read script from STDIN as no parameter is provided
// --> The script has to be run as remote script (content has to be transferred from CLI to gRPC server)
// after reading till EOF. Abort on timeout.
//
// 2) `P4wnP1_cli HID run --host "some hostname"` (host flag, no additional args)
// Assume that the CLI runs on a host different from the gRPC server, read script from STDIN as no parameter is provided
// --> The script has to be run as remote script (content has to be transferred from CLI to gRPC server)
// after reading till EOF. Abort on timeout.
//
// 3) `P4wnP1_cli HID run -c "Some HID script commands"`
// --> Same as 1) but content is read from `-c` instead of STDIN
//
// 4) `P4wnP1_cli HID run --host "some hostname" -c "Some HID script commands"`
// --> Same as 2) but content is read from `-c` instead of STDIN
//
// 5) `P4wnP1_cli HID run /some/path/to/file`
// Assume that the CLI runs on the same host as the server, assume that the given path refers to a file hosted on the
// gRPC server
// --> The script has to be run as local script (content is searched on P4wnP1 device by given path)
//
// 6) `P4wnP1_cli HID run --host "some hostname" -r /some/path/to/file`
// Assume that the CLI runs on a host different from the gRPC server and the given path points to a script on the host
// which runs the gRPC server (indicated by `-r`).
// --> The script has to be run as local script (content is searched on P4wnP1 device by given path)
//
// 7) `P4wnP1_cli HID run --host "some hostname" /some/path/to/file` (no host flag, no additional args)
// Assume that the CLI runs on a host different from the gRPC server and the given path points to a script on the host
// which runs the CLI client.
// --> The script has to be run as remote script (content has to be transferred from CLI to gRPC server)
// after reading it from the file.
//
// --> if flag `-c` is enabled, all following args are interpreted as script content
// --> if flag `-c` isn't set and args are present, the first arg is assumed to represent a filepath
// --> if flag `-c` isn't set and no additional args are provided, it is assumed that the script content has to be read
// from STDIN
//
// --> the two cases, which don't require a script content transfer from gRPC client to server:
// a) `-c` isn't set && --host == localhost && path is given as arg (example 5)
// b) `-c` isn't set && path is given with parameter `-r` (example 6)
//
// The logic above applies to both, running scripts synchronous with `run` or asynchronous with `job`


func parseHIDRunScriptCmd(cmd *cobra.Command, args []string) (serverScriptPath string, err error) {
	/*
	readFromStdin := false
	localFile := false //if true readFilePath refers to a file on the host of the rpcClient, else to a file on the rpcServer
	readFilePath := ""
	scriptContent :=""
	*/

	var srcReader io.Reader
	transferNeeded := false


	cFlagSet := cmd.Flags().ShorthandLookup("c").Changed
	rFlagSet := cmd.Flags().ShorthandLookup("n").Changed

	switch {
	case !rFlagSet && !cFlagSet:
		// if `-c` and `-r` aren't set and no additional args are provided, we have to read from STDIN
		if len(args) == 0 {
			//We have to read from STDIN
			srcReader = bufio.NewReader(os.Stdin)
			transferNeeded = true
		} else {
			// we assume the arg is a filePath
			if strings.ToLower(StrRemoteHost) != "localhost" {
				// file not hosted on RPC server, needs to be transferred
				transferNeeded=true
				f,err := os.OpenFile(args[0], os.O_RDONLY, os.ModePerm)
				if err != nil { return "",err }
				defer f.Close()
				srcReader = bufio.NewReader(f)
			} else {
				// assume RPC client is run from same host as RPC server and the script path refers to a local file
				transferNeeded = false
				serverScriptPath = common.PATH_HID_SCRIPTS + "/" + args[0]
			}
		}
	case rFlagSet:
		// the flag represents a script path on the RPC server, no matter where the RPC client is running, so we assume the script is already there
		transferNeeded = false
		serverScriptPath = common.PATH_HID_SCRIPTS + "/" + tmpRunStored
	case cFlagSet:
		// script content is provided by parameter and needs to be transferred
		transferNeeded = true
		srcReader = strings.NewReader(tmpHidCommands)
	case cFlagSet && rFlagSet:
		return "",errors.New("Couldn't use '-c' and '-r' at the same time")
	default:
		return "",errors.New("Invalid flag/parameter combination")
	}

	/*
	if readFromStdin {
		buf := make([]byte,1024)
		reader := bufio.NewReader(os.Stdin)
		for {
			n,rErr := reader.Read(buf)
			if rErr != nil {
				if rErr == io.EOF { break } else { return rErr }
			}
			chunk := buf[:n]
			scriptContent += string(chunk)
			fmt.Printf("Read %d bytes: %+q\n", n, string(chunk))
		}
	}


	fmt.Printf("readFromStdIn: %v path: %v content: %v\n", readFromStdin, readFilePath, scriptContent)
	*/

	if transferNeeded {
		// create random remote file

		serverScriptPath, err = ClientCreateTempFile(StrRemoteHost,StrRemotePort,"","HIDscript")
		if err != nil {
			return "",err
		} else {
			fmt.Printf("TempFile created: %s\n", serverScriptPath)
		}

		filename := filepath.Base(serverScriptPath)

		//transfer from reader to remote file
		err = ClientUploadFile(StrRemoteHost, StrRemotePort, srcReader, pb.AccessibleFolder_TMP, filename, true)
		if err != nil { return "",errors.New(fmt.Sprintf("Error transfering HIDScript content to P4wnP1 Server: %v", err))}
	}

	return
}

func cobraHidRun(cmd *cobra.Command, args []string) {
	serverScriptFilePath, err := parseHIDRunScriptCmd(cmd,args)
	if err != nil { log.Fatal(err)}

	ctx,cancel := context.WithCancel(context.Background())
	defer cancel()

	res,err := ClientHIDRunScript(StrRemoteHost, StrRemotePort, ctx, serverScriptFilePath, tmpHidTimeout)
	if err != nil { log.Fatal(err) }

	fmt.Printf("Result:\n%s\n", res.ResultJson)
	return
}


func cobraHidJob(cmd *cobra.Command, args []string) {
	serverScriptFilePath, err := parseHIDRunScriptCmd(cmd,args)
	if err != nil { log.Fatal(err)}


	job,err := ClientHIDRunScriptJob(StrRemoteHost, StrRemotePort, serverScriptFilePath, tmpHidTimeout)
	if err != nil { log.Fatal(err) }

	fmt.Printf("Job ID: %d\n", job.Id)
	return
}

func cobraHidJobCancel(cmd *cobra.Command, args []string) {
	if len(args) < 1 { log.Fatal("Job ID to cancel has to be given as argument\n")}
	jobID,err := strconv.ParseUint(args[0], 10, 32)
	if err != nil { log.Fatalf("Error parsing job ID '%s' ton integer\n", args[0])}

	err = ClientHIDCancelScriptJob(StrRemoteHost, StrRemotePort, uint32(jobID))
	if err != nil { log.Fatal(err) }

	return
}



func init() {
	rootCmd.AddCommand(hidCmd)
	hidCmd.AddCommand(hidRunCmd)
	hidCmd.AddCommand(hidJobCmd)
	hidJobCmd.AddCommand(hidJobCancelCmd)

	hidRunCmd.Flags().StringVarP(&tmpHidCommands, "commands","c", "", "HIDScript commands to run, given as string")
	hidRunCmd.Flags().StringVarP(&tmpRunStored, "name","n", "", "Run a stored HIDScript")
	hidRunCmd.Flags().Uint32VarP(&tmpHidTimeout, "timeout","t", 0, "Interrupt HIDScript after this timeout (seconds)")

	hidJobCmd.Flags().StringVarP(&tmpHidCommands, "commands","c", "", "HIDScript commands to run, given as string")
	hidJobCmd.Flags().StringVarP(&tmpRunStored, "name","n", "", "Run a stored HIDScript")
	hidJobCmd.Flags().Uint32VarP(&tmpHidTimeout, "timeout","t", 0, "Interrupt HIDScript after this timeout (seconds)")
}
