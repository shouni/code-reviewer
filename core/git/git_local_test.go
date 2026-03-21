package git

import "testing"

// [FIX] Minor: 実装側の仕様変更（リモートブランチ優先）に合わせてテストを修正
func TestBuildRefCandidates_PrefersBranchEvenForHexInput(t *testing.T) {
	// 16進数に見える入力であっても、まずは origin/ プレフィックス付きを優先する
	candidates := buildRefCandidates("f9211119e3")

	if len(candidates) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(candidates))
	}

	// 1番目の候補はリモートブランチであるべき
	if candidates[0].ref != "origin/f9211119e3" || !candidates[0].isBranchRef {
		t.Fatalf("unexpected first candidate (should be branch): %+v", candidates[0])
	}

	// 2番目の候補がコミットハッシュ（そのまま）であるべき
	if candidates[1].ref != "f9211119e3" || candidates[1].isBranchRef {
		t.Fatalf("unexpected second candidate (should be commit): %+v", candidates[1])
	}
}

func TestBuildRefCandidates_PreservesExplicitOriginRef(t *testing.T) {
	candidates := buildRefCandidates("origin/main")
	if len(candidates) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(candidates))
	}
	if candidates[0].ref != "origin/main" || !candidates[0].isBranchRef {
		t.Fatalf("unexpected candidate: %+v", candidates[0])
	}
}

func TestBuildRefCandidates_PrefersOriginForNamedBranch(t *testing.T) {
	candidates := buildRefCandidates("main")
	if len(candidates) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(candidates))
	}
	if candidates[0].ref != "origin/main" || !candidates[0].isBranchRef {
		t.Fatalf("unexpected first candidate: %+v", candidates[0])
	}
	if candidates[1].ref != "main" || candidates[1].isBranchRef {
		t.Fatalf("unexpected second candidate: %+v", candidates[1])
	}
}

func TestBranchNameFromResolvedRef(t *testing.T) {
	if got := branchNameFromResolvedRef("origin/release/v1"); got != "release/v1" {
		t.Fatalf("unexpected branch name: %s", got)
	}
	if got := branchNameFromResolvedRef("main"); got != "main" {
		t.Fatalf("unexpected branch name without origin prefix: %s", got)
	}
}
