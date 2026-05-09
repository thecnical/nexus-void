package social

import (
	"fmt"
	"strings"
)

// SocialManipulator is the AI-powered social engineering engine
type SocialManipulator struct {
	Target   string
	Persona  string
	Platform string
}

// SocialResult represents a social engineering finding
type SocialResult struct {
	Type       string  `json:"type"` // phishing, vishing, pretext, baiting
	Vector     string  `json:"vector"`
	Content    string  `json:"content"`
	Success    bool    `json:"success"`
	Confidence float64 `json:"confidence"`
}

func NewSocialManipulator(target, platform string) *SocialManipulator {
	return &SocialManipulator{
		Target:   target,
		Platform: platform,
		Persona:  "default",
	}
}

// GeneratePhishing generates AI-crafted phishing content
func (s *SocialManipulator) GeneratePhishing(template string, context map[string]string) string {
	fmt.Printf("[+] SOCIAL-MANIPULATOR generating phishing content for %s\n", s.Target)

	// AI-generated content based on target profile
	var content strings.Builder

	content.WriteString(fmt.Sprintf("Subject: %s\n\n", s.generateSubject(context)))
	content.WriteString(fmt.Sprintf("Dear %s,\n\n", context["name"]))
	content.WriteString(s.generateBody(context))
	content.WriteString("\n\nBest regards,\n")
	content.WriteString(s.generateSignature(context))

	return content.String()
}

func (s *SocialManipulator) generateSubject(context map[string]string) string {
	subjects := []string{
		"Urgent: Action Required on Your Account",
		"Security Alert: Unusual Login Detected",
		"Invoice Payment Confirmation Required",
		"Your Package Delivery Update",
		"Password Expiration Notice",
		"Document Shared With You",
		"Meeting Invitation: Action Required",
	}

	if s.Platform == "linkedin" {
		subjects = append(subjects, "New Connection Request")
	}

	return subjects[0] // Would use AI to pick best based on target
}

func (s *SocialManipulator) generateBody(context map[string]string) string {
	var body strings.Builder

	body.WriteString(fmt.Sprintf("We noticed unusual activity on your %s account. ", context["platform"]))
	body.WriteString("Please verify your identity by clicking the link below:\n\n")
	body.WriteString(fmt.Sprintf("http://%s.verify-account.com/secure/login\n\n", strings.ToLower(context["platform"])))
	body.WriteString("If you did not initiate this activity, please disregard this message.\n")

	return body.String()
}

func (s *SocialManipulator) generateSignature(context map[string]string) string {
	return fmt.Sprintf("IT Security Team\n%s Security Operations", context["company"])
}

// GenerateVishingScript generates voice phishing (vishing) script
func (s *SocialManipulator) GenerateVishingScript(scenario string) string {
	fmt.Printf("[+] SOCIAL-MANIPULATOR generating vishing script: %s\n", scenario)

	script := fmt.Sprintf(`VISHING SCRIPT - %s

INTRO:
"Hello, this is [NAME] calling from [COMPANY] IT Support. 
We have detected suspicious activity on your account."

BUILD RAPPORT:
- Confirm employee ID
- Mention recent company news
- Reference internal systems

PAYLOAD:
"We need to verify your credentials to secure your account.
Can you confirm your current password for our records?"

CLOSE:
"Thank you for your cooperation. We have secured your account.
You will receive a confirmation email shortly."

ESCALATION:
If target resists: "This is mandatory per company policy.
Your manager has been notified."
`, scenario)

	return script
}

// GeneratePretext creates a pretext scenario
func (s *SocialManipulator) GeneratePretext(role string) string {
	fmt.Printf("[+] SOCIAL-MANIPULATOR generating pretext as: %s\n", role)

	pretexts := map[string]string{
		"it_support": "You are an IT support technician performing emergency maintenance",
		"auditor":    "You are an external auditor conducting a security review",
		"vendor":     "You are a software vendor providing critical updates",
		"recruiter":  "You are a recruiter with an exclusive job opportunity",
		"journalist": "You are a journalist requesting an interview",
	}

	return pretexts[role]
}

// AnalyzeTargetProfile builds a psychological profile
func (s *SocialManipulator) AnalyzeTargetProfile(data map[string]string) map[string]interface{} {
	fmt.Printf("[+] SOCIAL-MANIPULATOR analyzing target profile for %s\n", s.Target)

	profile := make(map[string]interface{})

	// Analyze communication style
	if style, ok := data["communication_style"]; ok {
		profile["formality"] = s.assessFormality(style)
		profile["urgency_response"] = s.assessUrgencyResponse(style)
	}

	// Analyze technical sophistication
	if tech, ok := data["technical_level"]; ok {
		profile["technical_sophistication"] = tech
		profile["susceptibility"] = s.assessSusceptibility(tech)
	}

	// Determine best attack vector
	profile["recommended_vector"] = s.recommendVector(profile)
	profile["confidence"] = 0.85

	return profile
}

func (s *SocialManipulator) assessFormality(style string) string {
	if strings.Contains(strings.ToLower(style), "formal") {
		return "high"
	}
	return "medium"
}

func (s *SocialManipulator) assessUrgencyResponse(style string) float64 {
	if strings.Contains(strings.ToLower(style), "urgent") {
		return 0.9
	}
	return 0.6
}

func (s *SocialManipulator) assessSusceptibility(tech string) float64 {
	switch strings.ToLower(tech) {
	case "low":
		return 0.9
	case "medium":
		return 0.6
	case "high":
		return 0.3
	default:
		return 0.5
	}
}

func (s *SocialManipulator) recommendVector(profile map[string]interface{}) string {
	susceptibility, ok := profile["susceptibility"].(float64)
	if !ok {
		return "phishing"
	}

	if susceptibility > 0.8 {
		return "vishing"
	} else if susceptibility > 0.5 {
		return "phishing"
	}
	return "pretexting"
}

// GenerateBait creates USB drop bait content
func (s *SocialManipulator) GenerateBait(baitType string) string {
	fmt.Printf("[+] SOCIAL-MANIPULATOR generating %s bait\n", baitType)

	baits := map[string]string{
		"salary":       "Q3_Salary_Adjustments.xlsx.lnk",
		"layoff":       "Company_Restructuring_Notice.pdf.exe",
		"bonus":        "2024_Bonus_Payment_Details.docx.scr",
		"confidential": "Executive_Meeting_Notes_Confidential.doc.lnk",
		"resume":       "Resume_John_Doe_Senior_Engineer.pdf.exe",
	}

	return baits[baitType]
}
