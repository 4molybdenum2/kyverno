package webhooks

import (
	"strings"

	"github.com/nirmata/kyverno/pkg/annotations"
	"github.com/nirmata/kyverno/pkg/violation"

	"github.com/golang/glog"
	"github.com/nirmata/kyverno/pkg/event"
	"github.com/nirmata/kyverno/pkg/info"
)

//TODO: change validation from bool -> enum(validation, mutation)
func newEventInfoFromPolicyInfo(policyInfoList []*info.PolicyInfo, onUpdate bool, ruleType info.RuleType) ([]*event.Info, []*violation.Info) {
	var eventsInfo []*event.Info
	var violations []*violation.Info
	ok, msg := isAdmSuccesful(policyInfoList)
	// Some policies failed to apply succesfully
	if !ok {
		for _, pi := range policyInfoList {
			if pi.IsSuccessful() {
				continue
			}
			rules := pi.FailedRules()
			ruleNames := strings.Join(rules, ";")
			if !onUpdate {
				// CREATE
				eventsInfo = append(eventsInfo,
					event.NewEvent(policyKind, "", pi.Name, event.RequestBlocked, event.FPolicyApplyBlockCreate, pi.RName, ruleNames))

				glog.V(3).Infof("Rule(s) %s of policy %s blocked resource creation, error: %s\n", ruleNames, pi.Name, msg)
			} else {
				// UPDATE
				eventsInfo = append(eventsInfo,
					event.NewEvent(pi.RKind, pi.RNamespace, pi.RName, event.RequestBlocked, event.FPolicyApplyBlockUpdate, ruleNames, pi.Name))
				eventsInfo = append(eventsInfo,
					event.NewEvent(policyKind, "", pi.Name, event.RequestBlocked, event.FPolicyBlockResourceUpdate, pi.RName, ruleNames))
				glog.V(3).Infof("Request blocked events info has prepared for %s/%s and %s/%s\n", policyKind, pi.Name, pi.RKind, pi.RName)
			}
			// if report flag is set
			if pi.ValidationFailureAction == ReportViolation && ruleType == info.Validation {
				// Create Violations
				v := violation.BuldNewViolation(pi.Name, pi.RKind, pi.RNamespace, pi.RName, event.PolicyViolation.String(), pi.GetFailedRules())
				violations = append(violations, v)
			}
		}
	} else {
		if !onUpdate {
			// All policies were applied succesfully
			// CREATE
			for _, pi := range policyInfoList {
				rules := pi.SuccessfulRules()
				ruleNames := strings.Join(rules, ";")
				eventsInfo = append(eventsInfo,
					event.NewEvent(pi.RKind, pi.RNamespace, pi.RName, event.PolicyApplied, event.SRulesApply, ruleNames, pi.Name))

				glog.V(3).Infof("Success event info has prepared for %s/%s\n", pi.RKind, pi.RName)
			}
		}
	}
	return eventsInfo, violations
}

func addAnnotationsToResource(rawResource []byte, pi *info.PolicyInfo, ruleType info.RuleType) []byte {
	if len(pi.Rules) == 0 {
		return nil
	}
	// get annotations
	ann := annotations.ParseAnnotationsFromObject(rawResource)
	patch, err := annotations.PatchAnnotations(ann, pi, ruleType)
	if err != nil {
		glog.Error(err)
		return nil
	}
	return patch
}