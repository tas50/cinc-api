package cinc

import "context"

// ACE is a single access-control entry: the actors (users/clients by name)
// and groups granted one permission on one object.
type ACE struct {
	Actors []string `json:"actors"`
	Groups []string `json:"groups"`
}

// ACL is the complete permission set for a Chef object — five ACEs, one per
// standard Chef permission: create, read, update, delete, grant.
type ACL struct {
	Create ACE `json:"create"`
	Read   ACE `json:"read"`
	Update ACE `json:"update"`
	Delete ACE `json:"delete"`
	Grant  ACE `json:"grant"`
}

// ACLsService accesses the per-object ACL endpoints. Every Chef object that
// has identity (nodes, clients, groups, containers, cookbooks, data_bags,
// environments, policies, policy_groups, roles, ...) exposes an _acl
// subresource at the same path shape, so the object kind is a string.
type ACLsService struct{ client *Client }

// Get returns the full five-permission ACL of one object.
//
// objectType is the URL segment that identifies the kind ("nodes",
// "clients", "groups", "containers", "cookbooks", "data", "environments",
// "policies", "policy_groups", "roles", ...).
func (s *ACLsService) Get(ctx context.Context, objectType, name string) (*ACL, *Response, error) {
	a, resp, err := do[ACL](ctx, s.client, "GET",
		s.client.orgPath(objectType+"/"+name+"/_acl"), nil)
	return ptrOrNil(a, err), resp, err
}

// SetPermission rewrites one permission's ACE on one object. The Chef API
// requires the request body to wrap the new ACE under the permission name,
// e.g. {"update":{"actors":[],"groups":["admins"]}}.
//
// Nil Actors/Groups slices are coerced to empty arrays so the server does
// not reject the request for a null member list.
func (s *ACLsService) SetPermission(ctx context.Context, objectType, name, perm string, ace *ACE) error {
	body := map[string]any{
		perm: map[string]any{
			"actors": nonNil(ace.Actors),
			"groups": nonNil(ace.Groups),
		},
	}
	_, _, err := do[map[string]any](ctx, s.client, "PUT",
		s.client.orgPath(objectType+"/"+name+"/_acl/"+perm), body)
	return err
}
