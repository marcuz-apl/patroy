package main

import "testing"

func TestUserAgentFlagRegistered(t *testing.T) {
	f := rootCmd.Flags().Lookup("user-agent")
	if f == nil {
		t.Fatal("expected --user-agent flag to be registered on the root command")
	}
	if f.Usage == "" {
		t.Error("expected --user-agent flag to document its behavior")
	}
}
