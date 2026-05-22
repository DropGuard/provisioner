package main

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"github.com/spf13/cobra"
)

// Win32 API structures
type USER_INFO_1 struct {
	Name        *uint16
	Password    *uint16
	PasswordAge uint32
	Priv        uint32
	HomeDir     *uint16
	Comment     *uint16
	Flags       uint32
	ScriptPath  *uint16
}

type LOCALGROUP_MEMBERS_INFO_3 struct {
	DomainAndName *uint16
}

var Version = "dev"

var rootCmd = &cobra.Command{
	Use:           "user",
	Short:         "User provisioning tool",
	Long:          "Creates the daily-use Windows account (DailyUser) and assigns it to the local Administrators group natively using Win32 APIs.",
	Version:       Version,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runProvisioning()
	},
}

func isAdmin() bool {
	// Native Win32 check for Administrator privileges
	ret, _, _ := syscall.NewLazyDLL("shell32.dll").NewProc("IsUserAnAdmin").Call()
	return ret != 0
}

func runProvisioning() error {
	if !isAdmin() {
		return fmt.Errorf("Administrator privileges are required to run this command. Please run the terminal as Administrator.")
	}

	username := "DailyUser"

	netapi32 := syscall.NewLazyDLL("netapi32.dll")
	netUserAdd := netapi32.NewProc("NetUserAdd")
	netLocalGroupAddMembers := netapi32.NewProc("NetLocalGroupAddMembers")

	userPtr, err := syscall.UTF16PtrFromString(username)
	if err != nil {
		return fmt.Errorf("failed to process username: %w", err)
	}
	passPtr, _ := syscall.UTF16PtrFromString("") // No password

	ui := USER_INFO_1{
		Name:     userPtr,
		Password: passPtr,
		Priv:     1,               // USER_PRIV_USER
		Flags:    0x0040 | 0x0200, // UF_NORMAL_ACCOUNT | UF_SCRIPT
	}

	// Call NetUserAdd(servername=nil, level=1, buf, parm_err=nil)
	ret, _, _ := netUserAdd.Call(
		0,
		1,
		uintptr(unsafe.Pointer(&ui)),
		0,
	)

	// NERR_UserExists is 2224
	if ret != 0 {
		if ret == 2224 {
			fmt.Printf("[*] User '%s' already exists.\n", username)
		} else {
			return fmt.Errorf("failed to create user (error code %d): %v", ret, syscall.Errno(ret))
		}
	} else {
		fmt.Printf("[+] Created user '%s'.\n", username)
	}

	// Resolve the localized group name for Administrators (SID S-1-5-32-544)
	sidStr := "S-1-5-32-544"
	sid, err := syscall.StringToSid(sidStr)
	if err != nil {
		return fmt.Errorf("failed to convert SID: %w", err)
	}

	var nameLen, domainLen uint32 = 0, 0
	var sidUse uint32
	// Call once to get required buffer sizes
	err = syscall.LookupAccountSid(nil, sid, nil, &nameLen, nil, &domainLen, &sidUse)
	if err != nil && err != syscall.ERROR_INSUFFICIENT_BUFFER {
		return fmt.Errorf("failed to lookup group account sizes: %w", err)
	}

	nameBuf := make([]uint16, nameLen)
	domainBuf := make([]uint16, domainLen)
	err = syscall.LookupAccountSid(nil, sid, &nameBuf[0], &nameLen, &domainBuf[0], &domainLen, &sidUse)
	if err != nil {
		return fmt.Errorf("failed to lookup group account details: %w", err)
	}

	groupNameStr := syscall.UTF16ToString(nameBuf)

	groupNamePtr, err := syscall.UTF16PtrFromString(groupNameStr)
	if err != nil {
		return fmt.Errorf("failed to process group name: %w", err)
	}

	member := LOCALGROUP_MEMBERS_INFO_3{
		DomainAndName: userPtr,
	}

	// Call NetLocalGroupAddMembers(servername=nil, groupname, level=3, buf, totalentries=1)
	ret, _, _ = netLocalGroupAddMembers.Call(
		0,
		uintptr(unsafe.Pointer(groupNamePtr)),
		3,
		uintptr(unsafe.Pointer(&member)),
		1,
	)

	// ERROR_MEMBER_IN_ALIAS is 1378
	if ret != 0 {
		if ret == 1378 {
			fmt.Printf("[*] '%s' is already in the '%s' group.\n", username, groupNameStr)
		} else {
			return fmt.Errorf("failed to add user to group (error code %d): %v", ret, syscall.Errno(ret))
		}
	} else {
		fmt.Printf("[+] Added '%s' to the '%s' group.\n", username, groupNameStr)
	}

	fmt.Println("[+] Setup complete!")
	fmt.Println("======================================================================")
	fmt.Println("IMPORTANT:")
	fmt.Printf("Daily administrator account '%s' has been configured (no password).\n", username)
	fmt.Println("Please log off manually and switch to this account to use it:")
	fmt.Println("👉 Start -> Profile -> Sign out (Log off)")
	fmt.Println("======================================================================")
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "[-] Error: %v\n", err)
		os.Exit(1)
	}
}
