package bashguard

import "testing"

var owner = Policy{KubectlVerbs: []string{"exec"}}

func TestCheck(t *testing.T) {
	cases := []struct {
		command string
		policy  Policy
		allowed bool
	}{
		{"ls -la && git status", Policy{}, true},
		{"git commit -m 'deny kubectl apply'", Policy{}, true},
		{"grep -rn kubectl deploy/", Policy{}, true},
		{"aosguard ops kubectl get nodes", Policy{}, true},
		{"aosguard ops kubectl get nodes", owner, true},
		{"kubectl exec -n eco pod-0 -- ls /data", owner, true},
		{"kubectl --context kai-server -n eco exec pod-0 -- sh -c 'kubectl apply -f x'", owner, true},
		{"cd /tmp && kubectl exec pod-0 -- date", owner, true},
		{"bash -c 'kubectl exec pod-0 -- ls'", owner, true},
		{"echo $HOME", Policy{}, true},

		{"kubectl exec pod-0 -- ls", Policy{}, false},
		{"kubectl apply -f deploy.yaml", owner, false},
		{"kubectl -n eco apply -f deploy.yaml", owner, false},
		{"kubectl --as exec apply -f x", owner, false},
		{"kubectl --bogus exec pod-0", owner, false},
		{"/usr/local/bin/kubectl delete pod x", owner, false},
		{"k\"ube\"ctl apply -f x", owner, false},
		{"helm upgrade eco ./chart", owner, false},
		{"sudo -u root kubectl apply -f x", owner, false},
		{"env KUBECONFIG=/tmp/c kubectl apply -f x", owner, false},
		{"timeout 5s kubectl delete ns x", owner, false},
		{"xargs -I{} kubectl delete pod {}", owner, false},
		{"bash -c 'kubectl apply -f x'", owner, false},
		{"sh -ec \"helm uninstall eco\"", owner, false},
		{"eval kubectl apply -f x", owner, false},
		{"echo 'kubectl apply -f x' | bash", owner, false},
		{"bash <<'EOF'\nkubectl apply -f x\nEOF", owner, false},
		{"ssh kai@kai-server kubectl apply -f x", owner, false},
		{"python3 -c \"import os; os.system('kubectl apply')\"", owner, false},
		{"k=kubectl; $k apply -f x", owner, false},
		{"$(which kubectl) apply -f x", owner, false},
		{"f() { kubectl apply -f x; }; f", owner, false},
		{"kubectl", owner, false},
		{"echo ok; (cd /tmp && kubectl apply -f x)", owner, false},
	}
	for _, c := range cases {
		reason := Check(c.command, c.policy)
		if (reason == "") != c.allowed {
			t.Errorf("Check(%q, %v) = %q, want allowed=%v", c.command, c.policy.KubectlVerbs, reason, c.allowed)
		}
	}
}
