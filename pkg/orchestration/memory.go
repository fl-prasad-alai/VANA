// pkg/orchestration/memory.go

package orchestration

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	memoryMaxMessages = 80   // newest unsummarised messages read per refresh
	memoryMsgRunes    = 700  // each message is trimmed to this length
	memoryMaxRunes    = 1500 // hard cap on the stored summary
)

const memorySystemPrompt = `You maintain VANA's private memory of one person, so VANA can greet them like someone who remembers them.
You receive the existing memory (may be empty) and new conversation messages. Write the updated memory.

Write short plain-text notes in the third person, at most 120 words, no headings or markdown. Keep what still matters, merge the new, drop what is outdated. Cover only what was actually said:
- how they like to be called, and the language/script they write in
- life context (studies, work, family, where they are in life)
- what has been weighing on them, with rough timing ("early October: NEET exam stress")
- what helped and what did not; how they like VANA to respond (e.g. "wants to vent, not advice")
- progress and good news worth asking about next time

Never include: phone numbers, addresses, passwords, other people's full names, or details of self-harm methods. If they went through a crisis moment, note only "had a very hard moment on <date>; was given helplines" so VANA can check in gently.
Never invent anything. If the new messages add nothing worth remembering, return the existing memory unchanged.`

// RefreshMemory folds the user's new messages into their stored memory.
// It does nothing (and makes no AI call) when there are no new messages.
func (o *Orchestrator) RefreshMemory(ctx context.Context, userID string) (bool, error) {
	if o.db == nil {
		return false, nil
	}
	mem, err := o.db.GetUserMemory(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("load memory: %w", err)
	}
	msgs, err := o.db.GetMessagesSince(ctx, userID, mem.CoveredTo, memoryMaxMessages)
	if err != nil {
		return false, fmt.Errorf("load messages: %w", err)
	}
	if len(msgs) == 0 {
		return false, nil
	}

	var b strings.Builder
	fmt.Fprintf(&b, "EXISTING MEMORY:\n%s\n\nNEW MESSAGES (oldest first):\n", orNone(mem.Summary))
	lastConv := ""
	for _, m := range msgs {
		if m.ConversationID != lastConv {
			fmt.Fprintf(&b, "\n--- conversation on %s ---\n", m.CreatedAt.Format("2 Jan 2006"))
			lastConv = m.ConversationID
		}
		who := "Person"
		if m.Sender == "ai" {
			who = "VANA"
		}
		fmt.Fprintf(&b, "%s: %s\n", who, truncateRunes(strings.TrimSpace(m.Content), memoryMsgRunes))
	}
	b.WriteString("\nWrite the updated memory now.")

	ctxAI, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()
	summary, err := o.balancer.Summarize(ctxAI, memorySystemPrompt, b.String())
	if err != nil {
		return false, fmt.Errorf("summarise: %w", err)
	}
	summary = truncateRunes(strings.TrimSpace(stripMarkdown(summary)), memoryMaxRunes)
	if summary == "" {
		return false, nil
	}
	saved, err := o.db.SaveUserMemory(ctx, userID, summary, mem.CoveredTo, msgs[len(msgs)-1].CreatedAt)
	if err != nil {
		return false, fmt.Errorf("save memory: %w", err)
	}
	if saved {
		log.Printf("[VANA] memory refreshed for %s (%d new messages)", userID, len(msgs))
	}
	return saved, nil
}

// HasUnsummarised reports whether the user has messages not yet in memory
func (o *Orchestrator) HasUnsummarised(ctx context.Context, userID string) bool {
	if o.db == nil {
		return false
	}
	mem, err := o.db.GetUserMemory(ctx, userID)
	if err != nil {
		return false
	}
	n, err := o.db.CountMessagesSince(ctx, userID, mem.CoveredTo)
	return err == nil && n > 0
}

// memoryContext is the block added to every chat prompt
func (o *Orchestrator) memoryContext(ctx context.Context, userID string) string {
	if o.db == nil || userID == "" {
		return ""
	}
	mem, err := o.db.GetUserMemory(ctx, userID)
	if err != nil {
		return ""
	}
	var b strings.Builder
	if mem.FirstName != "" {
		fmt.Fprintf(&b, "PERSON'S NAME: %s (use it occasionally, not in every reply)\n", mem.FirstName)
	}
	if mem.Summary != "" {
		fmt.Fprintf(&b, "WHAT YOU REMEMBER FROM EARLIER CONVERSATIONS (you genuinely remember this; use it naturally, never recite it as a list; if they ask whether you remember them, say yes warmly and mention one or two relevant things):\n%s\n", mem.Summary)
	}
	return b.String()
}

func orNone(s string) string {
	if strings.TrimSpace(s) == "" {
		return "(none yet)"
	}
	return s
}

func truncateRunes(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n]) + "…"
}

// stripMarkdown removes heading/bullet markers a model may add despite instructions
func stripMarkdown(s string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		l = strings.TrimLeft(l, "#*-• ")
		lines[i] = strings.ReplaceAll(l, "**", "")
	}
	return strings.Join(lines, "\n")
}
