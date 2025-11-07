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

        // Check if user can create projects (superuser or is project_owner in any project)
        const canCreateProject = () => {
            if (superuser) return true;
            const pr = user?.project_roles;
            if (!pr) return false;
            // Check if user is project_owner in any project
            for (const roleKey of Object.values(pr)) {
                if (roleKey === 'project_owner') {
                    return true;
                }
            }
            return false;
        };

        // Check if user can delete a specific project
        const canDeleteProject = (projId?: string | number) => {
            if (superuser) return true;
            // Use projectId from hook if projId not provided
            const targetProjectId = projId ?? projectId;
            if (!targetProjectId || !pp) return false;
            // Check if user has project.manage permission in this specific project
            const perms = pp[String(targetProjectId)];
            return perms && perms.includes(PERMISSIONS.project.manage);
        };

        // Check if user can manage a specific project
        const canManageProject = (projId?: string | number) => {
            if (superuser) return true;
            // Use projectId from hook if projId not provided
            const targetProjectId = projId ?? projectId;
            if (!targetProjectId || !pp) return false;
            // Check if user has project.manage permission in this specific project
            const perms = pp[String(targetProjectId)];
            return perms && perms.includes(PERMISSIONS.project.manage);
        };

        return {
            isSuperuser: superuser,
            has: check,
            canViewProject: () => check(PERMISSIONS.project.view),
            canManageProject,
            canViewAudit: () => check(PERMISSIONS.audit.view),
            canManageMembership: () => check(PERMISSIONS.membership.manage),
            canCreateProject,
            canDeleteProject,
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
