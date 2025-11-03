import { useMemo } from 'react';
import { useAuth } from './AuthContext';

// Permission keys as constants to avoid typos
export const PERMISSIONS = {
    project: {
        view: 'project.view',
        manage: 'project.manage',
    },
    audit: {
        view: 'audit.view',
    },
    membership: {
        manage: 'membership.manage',
    },
} as const;

export type PermissionKey =
    | typeof PERMISSIONS.project.view
    | typeof PERMISSIONS.project.manage
    | typeof PERMISSIONS.audit.view
    | typeof PERMISSIONS.membership.manage

export function hasPermission(
    projectId: string | number | undefined,
    perm: PermissionKey,
    opts?: { isSuperuser?: boolean; projectPermissions?: Record<string, string[]> | undefined },
): boolean {
    const isSuperuser = opts?.isSuperuser ?? false;
    if (isSuperuser) return true;

    // For global permissions (like category.manage), check if user has it on any project

    // For project-specific permissions, we need projectId
    if (!projectId) return false;

    const pp = opts?.projectPermissions;
    if (!pp) return false;

    const perms = pp[String(projectId)];
    if (!perms || perms.length === 0) return false;

    return perms.includes(perm);
}

export function useRBAC(projectId?: string | number) {
    const { user } = useAuth();

    const guards = useMemo(() => {
        const superuser = Boolean(user?.is_superuser);
        const pp = user?.project_permissions;

        const check = (p: PermissionKey) => hasPermission(projectId, p, { isSuperuser: superuser, projectPermissions: pp });

        return {
            isSuperuser: superuser,
            has: check,
            canViewProject: () => check(PERMISSIONS.project.view),
            canManageProject: () => check(PERMISSIONS.project.manage),
            canViewAudit: () => check(PERMISSIONS.audit.view),
            canManageMembership: () => check(PERMISSIONS.membership.manage),
        };
    }, [user, projectId]);

    return guards;
}

// Standalone helpers mirroring the spec (for modules that cannot use hooks)
export const Guard = {
    canViewProject(projectId: string | number, isSuperuser?: boolean, projectPermissions?: Record<string, string[]>) {
        return hasPermission(projectId, PERMISSIONS.project.view, { isSuperuser, projectPermissions });
    },
    canManageProject(projectId: string | number, isSuperuser?: boolean, projectPermissions?: Record<string, string[]>) {
        return hasPermission(projectId, PERMISSIONS.project.manage, { isSuperuser, projectPermissions });
    },
    canViewAudit(projectId: string | number, isSuperuser?: boolean, projectPermissions?: Record<string, string[]>) {
        return hasPermission(projectId, PERMISSIONS.audit.view, { isSuperuser, projectPermissions });
    },
    canManageMembership(projectId: string | number, isSuperuser?: boolean, projectPermissions?: Record<string, string[]>) {
        return hasPermission(projectId, PERMISSIONS.membership.manage, { isSuperuser, projectPermissions });
    },
};
