import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { roleLabels } from "../entities/labels";
import type { UserRole } from "../entities/types";
import { api } from "../shared/api/endpoints";
import { formatDateTime } from "../shared/lib/format";
import { Panel } from "../shared/ui/Panel";
import { RoleBadge } from "../shared/ui/StatusBadge";

const roles: UserRole[] = ["admin", "analyst", "operator", "viewer"];

export function UsersPage() {
  const queryClient = useQueryClient();
  const { data = [] } = useQuery({ queryKey: ["users"], queryFn: api.users });
  const mutation = useMutation({
    mutationFn: ({ id, role }: { id: string; role: UserRole }) => api.updateUserRole(id, role),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["users"] })
  });

  return (
    <div className="page-grid">
      <Panel title="Users and access">
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>User</th>
                <th>Login</th>
                <th>Email</th>
                <th>Current role</th>
                <th>Status</th>
                <th>Change role</th>
                <th>Last seen</th>
              </tr>
            </thead>
            <tbody>
              {data.map((user) => (
                <tr key={user.id}>
                  <td>{user.name}</td>
                  <td>{user.login}</td>
                  <td>{user.email}</td>
                  <td><RoleBadge role={user.role} /></td>
                  <td>{user.isActive ? "Active" : "Inactive"}</td>
                  <td>
                    <select
                      value={user.role}
                      onChange={(event) => mutation.mutate({ id: user.id, role: event.target.value as UserRole })}
                    >
                      {roles.map((role) => (
                        <option key={role} value={role}>
                          {roleLabels[role]}
                        </option>
                      ))}
                    </select>
                  </td>
                  <td>{formatDateTime(user.lastSeen)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Panel>
    </div>
  );
}
