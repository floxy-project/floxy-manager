import React from 'react';
import { LDAPSyncProgress } from '../utils/api';

interface LDAPSyncProgressBarProps {
  isRunning: boolean;
  progress: number; // 0..100 (percentage)
  currentStep?: string;
  processedItems?: number;
  totalItems?: number;
  estimatedTime?: string;
  startTime?: string;
}

const LDAPSyncProgressBar: React.FC<LDAPSyncProgressBarProps> = ({
  isRunning,
  progress,
  currentStep,
  processedItems,
  totalItems,
  estimatedTime,
  startTime,
}) => {
  if (!isRunning) return null;

  const percent = Math.round(progress || 0);

  return (
    <div className="mb-6 p-4 rounded-lg border-2 border-slate-900 dark:border-[#ff6b35] bg-white dark:bg-[#1e1e1e] max-w-md mx-auto">
      <h3 className="text-lg font-semibold mb-3 text-slate-900 dark:text-[#ff6b35]">
        LDAP Users Synchronization
      </h3>
      
      <div className="flex items-center gap-4 mb-3">
        <div className="flex-1">
          <div className="w-full bg-slate-200 dark:bg-[#3e3e42] rounded-full h-2.5">
            <div
              className="bg-slate-900 dark:bg-[#ff6b35] h-2.5 rounded-full transition-all duration-300"
              style={{ width: `${percent}%` }}
            />
          </div>
        </div>
        <span className="text-sm font-medium text-slate-600 dark:text-[#ff4500] min-w-[40px]">
          {percent}%
        </span>
      </div>

      <div className="mt-2 space-y-1">
        {currentStep && (
          <p className="text-sm text-slate-700 dark:text-slate-300">
            Step: {currentStep}
          </p>
        )}
        {typeof processedItems === 'number' && typeof totalItems === 'number' && (
          <p className="text-sm text-slate-600 dark:text-[#ff4500]">
            Processed: {processedItems} of {totalItems}
          </p>
        )}
        {estimatedTime && (
          <p className="text-sm text-slate-600 dark:text-[#ff4500]">
            Estimated time remaining: {estimatedTime}
          </p>
        )}
        {startTime && (
          <p className="text-xs text-slate-500 dark:text-slate-400">
            Started: {new Date(startTime).toLocaleString()}
          </p>
        )}
      </div>
    </div>
  );
};

export default LDAPSyncProgressBar;

