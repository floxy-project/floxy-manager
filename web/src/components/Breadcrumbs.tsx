import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { ChevronRight } from 'lucide-react';

interface BreadcrumbItem {
  label: string;
  path: string;
}

export const Breadcrumbs: React.FC = () => {
  const location = useLocation();
  const pathname = location.pathname;

  const buildBreadcrumbs = (): BreadcrumbItem[] => {
    const items: BreadcrumbItem[] = [];
    const pathParts = pathname.split('/').filter(Boolean);

    // Always start with Dashboard (root)
    items.push({ label: 'Dashboard', path: '/' });

    // Handle /tenants
    if (pathParts[0] === 'tenants') {
      items.push({ label: 'Tenants', path: '/tenants' });

      // Handle /tenants/:tenantId
      if (pathParts[1]) {
        const tenantId = pathParts[1];
        
        // Handle /tenants/:tenantId/projects
        if (pathParts[2] === 'projects') {
          // Handle /tenants/:tenantId/projects/:projectId
          if (pathParts[3]) {
            const projectId = pathParts[3];
            items.push({ 
              label: `Project ${projectId}`, 
              path: `/tenants/${tenantId}/projects/${projectId}` 
            });

            // Handle nested routes under project
            const remainingPath = pathParts.slice(4);
            if (remainingPath.length > 0) {
              const route = remainingPath[0];
              const basePath = `/tenants/${tenantId}/projects/${projectId}`;

              // Map route names to labels
              const routeLabels: Record<string, string> = {
                'dashboard': 'Dashboard',
                'workflows': 'Workflows',
                'instances': 'Instances',
                'stats': 'Statistics',
                'dlq': 'Dead Letter Queue',
              };

              const routeLabel = routeLabels[route] || route;
              
              // For detail pages, add the list page first
              if (remainingPath.length > 1) {
                items.push({ 
                  label: routeLabel, 
                  path: `${basePath}/${route}` 
                });
                
                // Add detail page
                const detailId = remainingPath[1];
                const detailLabel = route === 'workflows' 
                  ? `Workflow ${detailId}`
                  : route === 'instances'
                  ? `Instance ${detailId}`
                  : route === 'dlq'
                  ? `DLQ Item ${detailId}`
                  : detailId;
                
                items.push({ 
                  label: detailLabel, 
                  path: `${basePath}/${route}/${detailId}` 
                });
              } else {
                // Just the list page
                items.push({ 
                  label: routeLabel, 
                  path: `${basePath}/${route}` 
                });
              }
            }
          } else {
            // /tenants/:tenantId/projects - list of projects
            items.push({ 
              label: 'Projects', 
              path: `/tenants/${tenantId}/projects` 
            });
          }
        } else {
          // /tenants/:tenantId (without projects) - shouldn't happen in current routing
          items.push({ 
            label: `Tenant ${tenantId}`, 
            path: `/tenants/${tenantId}` 
          });
        }
      }
    } else if (pathParts.length > 0) {
      // Handle other root routes
      const routeLabels: Record<string, string> = {
        'workflows': 'Workflows',
        'instances': 'Instances',
        'stats': 'Statistics',
        'dlq': 'Dead Letter Queue',
      };

      pathParts.forEach((part, index) => {
        const path = '/' + pathParts.slice(0, index + 1).join('/');
        const label = routeLabels[part] || part.charAt(0).toUpperCase() + part.slice(1);
        items.push({ label, path });
      });
    }

    return items;
  };

  const breadcrumbs = buildBreadcrumbs();

  // Don't show breadcrumbs on root pages (Dashboard, Tenants list)
  if (breadcrumbs.length <= 2) {
    return null;
  }

  return (
    <nav className="mb-6" aria-label="Breadcrumb">
      <ol className="flex items-center gap-2 text-sm flex-wrap">
        {breadcrumbs.map((crumb, index) => {
          const isLast = index === breadcrumbs.length - 1;
          
          return (
            <li key={`${crumb.path}-${index}`} className="flex items-center gap-2">
              {index > 0 && (
                <ChevronRight className="w-4 h-4 text-slate-400 flex-shrink-0" />
              )}
              {isLast ? (
                <span className="text-slate-900 dark:text-[#ff6b35] font-medium">
                  {crumb.label}
                </span>
              ) : (
                <Link
                  to={crumb.path}
                  className="text-slate-600 dark:text-[#ff4500] hover:text-slate-900 dark:hover:text-[#ff6b35] transition-colors"
                >
                  {crumb.label}
                </Link>
              )}
            </li>
          );
        })}
      </ol>
    </nav>
  );
};

