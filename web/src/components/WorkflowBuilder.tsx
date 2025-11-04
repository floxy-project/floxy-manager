import React, { useState, useMemo, useEffect } from 'react';
import { WorkflowGraph, type GraphDefinition, type StepDefinition } from './WorkflowGraph';

export interface WorkflowDefinition {
  name: string;
  version: number;
  definition: {
    start: string;
    steps: Record<string, StepDefinition>;
    dlq_enabled: boolean;
  };
}

interface WorkflowBuilderProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: (definition: WorkflowDefinition) => void;
  initialDefinition?: WorkflowDefinition;
}

export const WorkflowBuilder: React.FC<WorkflowBuilderProps> = ({
  isOpen,
  onClose,
  onSave,
  initialDefinition,
}) => {
  const [workflowName, setWorkflowName] = useState(initialDefinition?.name || '');
  const [workflowVersion, setWorkflowVersion] = useState(initialDefinition?.version || 1);
  const [dlqEnabled, setDlqEnabled] = useState(initialDefinition?.definition.dlq_enabled || false);
  const [steps, setSteps] = useState<Record<string, StepDefinition>>(
    initialDefinition?.definition.steps || {}
  );
  const [startStep, setStartStep] = useState(initialDefinition?.definition.start || '');
  const [selectedStep, setSelectedStep] = useState<string | null>(null);

  // Clear state when modal is closed
  useEffect(() => {
    if (!isOpen) {
      setWorkflowName('');
      setWorkflowVersion(1);
      setDlqEnabled(false);
      setSteps({});
      setStartStep('');
      setSelectedStep(null);
    }
  }, [isOpen]);

  const handleAddStep = (type: StepDefinition['type']) => {
    const stepName = `step_${Date.now()}`;
    const newStep: StepDefinition = {
      name: stepName,
      type,
      next: [],
      metadata: {},
      max_retries: 3,
      no_idempotent: false,
    };

    if (type === 'task') {
      newStep.handler = '';
    } else if (type === 'condition') {
      newStep.condition = '';
    } else if (type === 'join') {
      newStep.wait_for = [];
      newStep.join_strategy = 'all';
    } else if (type === 'fork' || type === 'parallel') {
      newStep.parallel = [];
    }

    if (startStep === '') {
      // First step - set as start
      newStep.prev = '_root_';
      setStartStep(stepName);
      setSteps({ ...steps, [stepName]: newStep });
    } else {
      // Try to connect to the currently selected step or last step
      const connectToStep = selectedStep || Object.keys(steps)[Object.keys(steps).length - 1];
      if (connectToStep && steps[connectToStep]) {
        const connectStep = steps[connectToStep];
        if (!connectStep.next) {
          connectStep.next = [];
        }
        if (!connectStep.next.includes(stepName)) {
          connectStep.next.push(stepName);
        }
        newStep.prev = connectToStep;
        setSteps({ ...steps, [connectToStep]: connectStep, [stepName]: newStep });
      } else {
        setSteps({ ...steps, [stepName]: newStep });
      }
    }

    setSelectedStep(stepName);
  };

  const handleDeleteStep = (stepName: string) => {
    const newSteps = { ...steps };
    delete newSteps[stepName];

    // Remove references
    Object.values(newSteps).forEach(step => {
      if (step.next) {
        step.next = step.next.filter(n => n !== stepName);
      }
      if (step.prev === stepName) {
        step.prev = '';
      }
      if (step.else === stepName) {
        step.else = undefined;
      }
      if (step.on_failure === stepName) {
        step.on_failure = undefined;
      }
      if (step.parallel) {
        step.parallel = step.parallel.filter(p => p !== stepName);
      }
      if (step.wait_for) {
        step.wait_for = step.wait_for.filter(w => w !== stepName);
      }
    });

    setSteps(newSteps);

    if (startStep === stepName) {
      const remainingSteps = Object.keys(newSteps);
      setStartStep(remainingSteps.length > 0 ? remainingSteps[0] : '');
    }
  };

  const handleConnectSteps = (from: string, to: string) => {
    const fromStep = steps[from];
    if (!fromStep) return;

    const newSteps = { ...steps };
    if (!newSteps[from].next) {
      newSteps[from].next = [];
    }
    if (!newSteps[from].next.includes(to)) {
      newSteps[from].next.push(to);
    }
    newSteps[to].prev = from;
    setSteps(newSteps);
  };

  const handleUpdateStep = (stepName: string, updates: Partial<StepDefinition>) => {
    setSteps({
      ...steps,
      [stepName]: { ...steps[stepName], ...updates },
    });
  };

  const handleSave = () => {
    if (!workflowName.trim()) {
      alert('Please enter workflow name');
      return;
    }

    if (startStep === '') {
      alert('Please add at least one step');
      return;
    }

    const definition: WorkflowDefinition = {
      name: workflowName,
      version: workflowVersion,
      definition: {
        start: startStep,
        steps: steps,
        dlq_enabled: dlqEnabled,
      },
    };

    onSave(definition);
  };

  const graphDefinition: GraphDefinition = useMemo(() => ({
    start: startStep,
    steps: steps,
  }), [startStep, steps]);

  const exportJSON = () => {
    // Clean up steps to match Go builder format
    const cleanedSteps: Record<string, any> = {};
    
    Object.entries(steps).forEach(([name, step]) => {
      const cleaned: any = {
        name: step.name,
        type: step.type,
        delay: 0,
        handler: step.handler || '',
        timeout: 0,
        metadata: step.metadata || {},
        parallel: step.parallel || null,
        wait_for: step.wait_for || null,
        condition: step.condition || '',
        on_failure: step.on_failure || '',
        max_retries: step.max_retries || 0,
        retry_delay: 0,
        join_strategy: step.join_strategy || '',
        no_idempotent: step.no_idempotent || false,
        retry_strategy: 0,
      };

      // Add optional fields
      if (step.next && step.next.length > 0) {
        cleaned.next = step.next;
      } else {
        cleaned.next = step.type === 'task' && !step.next ? [] : null;
      }
      
      if (step.prev) {
        // If this is the start step, prev should be "_root_"
        cleaned.prev = (name === startStep) ? '_root_' : step.prev;
      } else {
        // If no prev and it's the start step, set to "_root_"
        cleaned.prev = (name === startStep) ? '_root_' : '';
      }

      if (step.else) {
        cleaned.else = step.else;
      }

      // Clean up null arrays
      if (cleaned.parallel === null || cleaned.parallel.length === 0) {
        cleaned.parallel = null;
      }
      if (cleaned.wait_for === null || cleaned.wait_for.length === 0) {
        cleaned.wait_for = null;
      }

      cleanedSteps[name] = cleaned;
    });

    const definition = {
      start: startStep,
      steps: cleanedSteps,
      dlq_enabled: dlqEnabled,
    };

    const jsonStr = JSON.stringify(definition, null, 2);
    const blob = new Blob([jsonStr], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${workflowName || 'workflow'}-v${workflowVersion}.json`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  if (!isOpen) return null;

  return (
    <div 
      className="fixed inset-0 z-[9999] flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm"
      onClick={onClose}
    >
      <div 
        className="bg-white dark:bg-[#252526] rounded-xl shadow-2xl max-w-[95vw] w-full max-h-[90vh] overflow-hidden flex flex-col relative border border-slate-200 dark:border-[#3e3e42]"
        onClick={(e) => e.stopPropagation()}
        style={{ width: '1400px', maxWidth: '95vw' }}
      >
        {/* Header */}
        <div className="flex items-start justify-between px-6 pt-6 pb-4 border-b border-slate-200">
          <div className="flex-1">
            <h3 className="text-xl font-bold text-slate-900">Workflow Builder</h3>
            <div className="mt-4 flex gap-4">
              <div className="flex-1">
                <label className="block text-sm font-medium text-slate-700 mb-1">
                  Workflow Name
                </label>
                <input
                  type="text"
                  value={workflowName}
                  onChange={(e) => setWorkflowName(e.target.value)}
                  className="w-full px-3 py-2 border border-slate-300 rounded-lg"
                  placeholder="my-workflow"
                />
              </div>
              <div className="w-24">
                <label className="block text-sm font-medium text-slate-700 mb-1">
                  Version
                </label>
                <input
                  type="number"
                  value={workflowVersion}
                  onChange={(e) => setWorkflowVersion(parseInt(e.target.value) || 1)}
                  className="w-full px-3 py-2 border border-slate-300 rounded-lg"
                  min="1"
                />
              </div>
              <div className="flex items-end">
                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={dlqEnabled}
                    onChange={(e) => setDlqEnabled(e.target.checked)}
                    className="rounded"
                  />
                  <span className="text-sm text-slate-700">DLQ Enabled</span>
                </label>
              </div>
            </div>
          </div>
          <button
            className="ml-4 w-8 h-8 flex items-center justify-center bg-slate-100 hover:bg-slate-200 rounded-lg transition-colors text-slate-600 hover:text-slate-900 text-xl leading-none"
            onClick={onClose}
          >
            ×
          </button>
        </div>

        {/* Main Content */}
        <div className="flex-1 flex overflow-hidden">
          {/* Sidebar - Step Types */}
          <div className="w-64 border-r border-slate-200 p-4 overflow-y-auto">
            <h4 className="font-semibold text-slate-900 mb-3">Add Step</h4>
            <div className="space-y-2">
              <button
                className="w-full text-left px-3 py-2 bg-blue-50 hover:bg-blue-100 rounded-lg text-sm"
                onClick={() => handleAddStep('task')}
              >
                📝 Task
              </button>
              <button
                className="w-full text-left px-3 py-2 bg-green-50 hover:bg-green-100 rounded-lg text-sm"
                onClick={() => handleAddStep('condition')}
              >
                🔀 Condition
              </button>
              <button
                className="w-full text-left px-3 py-2 bg-purple-50 hover:bg-purple-100 rounded-lg text-sm"
                onClick={() => handleAddStep('fork')}
              >
                🍴 Fork
              </button>
              <button
                className="w-full text-left px-3 py-2 bg-yellow-50 hover:bg-yellow-100 rounded-lg text-sm"
                onClick={() => handleAddStep('join')}
              >
                🔗 Join
              </button>
              <button
                className="w-full text-left px-3 py-2 bg-orange-50 hover:bg-orange-100 rounded-lg text-sm"
                onClick={() => handleAddStep('parallel')}
              >
                ⚡ Parallel
              </button>
              <button
                className="w-full text-left px-3 py-2 bg-pink-50 hover:bg-pink-100 rounded-lg text-sm"
                onClick={() => handleAddStep('human')}
              >
                👤 Human Decision
              </button>
              <button
                className="w-full text-left px-3 py-2 bg-indigo-50 hover:bg-indigo-100 rounded-lg text-sm"
                onClick={() => handleAddStep('save_point')}
              >
                💾 Save Point
              </button>
            </div>

            <div className="mt-6">
              <h4 className="font-semibold text-slate-900 mb-3">Steps</h4>
              <div className="space-y-1">
                {Object.keys(steps).map((stepName) => (
                  <div
                    key={stepName}
                    className={`flex items-center justify-between px-2 py-1 rounded cursor-pointer ${
                      selectedStep === stepName ? 'bg-blue-100' : 'hover:bg-slate-100'
                    }`}
                    onClick={() => setSelectedStep(stepName)}
                  >
                    <span className="text-sm">{stepName}</span>
                    <button
                      className="text-red-500 hover:text-red-700 text-xs"
                      onClick={(e) => {
                        e.stopPropagation();
                        if (confirm(`Delete step "${stepName}"?`)) {
                          handleDeleteStep(stepName);
                        }
                      }}
                    >
                      ×
                    </button>
                  </div>
                ))}
                {Object.keys(steps).length === 0 && (
                  <p className="text-xs text-slate-500">No steps yet</p>
                )}
              </div>
            </div>
          </div>

          {/* Graph View */}
          <div className="flex-1 overflow-auto p-4">
            {Object.keys(steps).length === 0 ? (
              <div className="flex items-center justify-center h-full text-slate-500">
                <div className="text-center">
                  <p className="text-lg mb-2">No steps added yet</p>
                  <p className="text-sm">Add a step from the sidebar to get started</p>
                </div>
              </div>
            ) : (
              <WorkflowGraph definition={graphDefinition} />
            )}
          </div>

          {/* Step Editor Panel */}
          {selectedStep && steps[selectedStep] && (
            <div className="w-80 border-l border-slate-200 p-4 overflow-y-auto">
              <div className="flex items-center justify-between mb-4">
                <h4 className="font-semibold text-slate-900">Edit Step</h4>
                <button
                  className="text-slate-500 hover:text-slate-700"
                  onClick={() => setSelectedStep(null)}
                >
                  ×
                </button>
              </div>
              <StepEditor
                step={steps[selectedStep]}
                allSteps={Object.keys(steps)}
                onUpdate={(updates) => handleUpdateStep(selectedStep, updates)}
              />
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="px-6 py-4 border-t border-slate-200 flex justify-between items-center">
          <button
            className="btn btn-secondary"
            onClick={exportJSON}
          >
            Export JSON
          </button>
          <div className="flex gap-3">
            <button
              className="btn btn-secondary"
              onClick={onClose}
            >
              Cancel
            </button>
            <button
              className="btn btn-primary"
              onClick={handleSave}
            >
              Save
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

interface StepEditorProps {
  step: StepDefinition;
  allSteps: string[];
  onUpdate: (updates: Partial<StepDefinition>) => void;
}

const StepEditor: React.FC<StepEditorProps> = ({ step, allSteps, onUpdate }) => {
  return (
    <div className="space-y-4">
      <div>
        <label className="block text-sm font-medium text-slate-700 mb-1">Name</label>
        <input
          type="text"
          value={step.name}
          onChange={(e) => onUpdate({ name: e.target.value })}
          className="w-full px-3 py-2 border border-slate-300 rounded-lg"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-slate-700 mb-1">Type</label>
        <div className="px-3 py-2 bg-slate-100 rounded-lg text-sm">{step.type}</div>
      </div>

      {step.type === 'task' && (
        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1">Handler</label>
          <input
            type="text"
            value={step.handler || ''}
            onChange={(e) => onUpdate({ handler: e.target.value })}
            className="w-full px-3 py-2 border border-slate-300 rounded-lg"
            placeholder="handler-name"
          />
        </div>
      )}

      {step.type === 'condition' && (
        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1">Condition</label>
          <textarea
            value={step.condition || ''}
            onChange={(e) => onUpdate({ condition: e.target.value })}
            className="w-full px-3 py-2 border border-slate-300 rounded-lg font-mono text-sm"
            rows={3}
            placeholder='{{ gt .count 5 }}'
          />
        </div>
      )}

      <div>
        <label className="block text-sm font-medium text-slate-700 mb-1">Max Retries</label>
        <input
          type="number"
          value={step.max_retries || 3}
          onChange={(e) => onUpdate({ max_retries: parseInt(e.target.value) || 0 })}
          className="w-full px-3 py-2 border border-slate-300 rounded-lg"
          min="0"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-slate-700 mb-1">Next Steps</label>
        <div className="space-y-1">
          {step.next?.map((nextStep, idx) => (
            <div key={idx} className="flex items-center gap-2">
              <select
                value={nextStep}
                onChange={(e) => {
                  const newNext = [...(step.next || [])];
                  newNext[idx] = e.target.value;
                  onUpdate({ next: newNext });
                }}
                className="flex-1 px-3 py-2 border border-slate-300 rounded-lg text-sm"
              >
                <option value="">Select step</option>
                {allSteps.filter(s => s !== step.name).map(s => (
                  <option key={s} value={s}>{s}</option>
                ))}
              </select>
              <button
                className="text-red-500 hover:text-red-700"
                onClick={() => {
                  const newNext = step.next?.filter((_, i) => i !== idx) || [];
                  onUpdate({ next: newNext });
                }}
              >
                ×
              </button>
            </div>
          ))}
          <button
            className="text-sm text-blue-600 hover:text-blue-800"
            onClick={() => {
              const newNext = [...(step.next || []), ''];
              onUpdate({ next: newNext });
            }}
          >
            + Add Next Step
          </button>
        </div>
      </div>

      {step.type !== 'join' && step.type !== 'condition' && (
        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1">On Failure</label>
          <select
            value={step.on_failure || ''}
            onChange={(e) => onUpdate({ on_failure: e.target.value || undefined })}
            className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm"
          >
            <option value="">None</option>
            {allSteps.filter(s => s !== step.name).map(s => (
              <option key={s} value={s}>{s}</option>
            ))}
          </select>
        </div>
      )}

      {step.type === 'condition' && (
        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1">Else Step</label>
          <select
            value={step.else || ''}
            onChange={(e) => onUpdate({ else: e.target.value || undefined })}
            className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm"
          >
            <option value="">None</option>
            {allSteps.filter(s => s !== step.name).map(s => (
              <option key={s} value={s}>{s}</option>
            ))}
          </select>
        </div>
      )}

      {step.type === 'join' && (
        <>
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">Join Strategy</label>
            <select
              value={step.join_strategy || 'all'}
              onChange={(e) => onUpdate({ join_strategy: e.target.value as 'all' | 'any' })}
              className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm"
            >
              <option value="all">All</option>
              <option value="any">Any</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">Wait For Steps</label>
            <div className="space-y-1">
              {step.wait_for?.map((waitStep, idx) => (
                <div key={idx} className="flex items-center gap-2">
                  <select
                    value={waitStep}
                    onChange={(e) => {
                      const newWaitFor = [...(step.wait_for || [])];
                      newWaitFor[idx] = e.target.value;
                      onUpdate({ wait_for: newWaitFor });
                    }}
                    className="flex-1 px-3 py-2 border border-slate-300 rounded-lg text-sm"
                  >
                    <option value="">Select step</option>
                    {allSteps.filter(s => s !== step.name).map(s => (
                      <option key={s} value={s}>{s}</option>
                    ))}
                  </select>
                  <button
                    className="text-red-500 hover:text-red-700"
                    onClick={() => {
                      const newWaitFor = step.wait_for?.filter((_, i) => i !== idx) || [];
                      onUpdate({ wait_for: newWaitFor });
                    }}
                  >
                    ×
                  </button>
                </div>
              ))}
              <button
                className="text-sm text-blue-600 hover:text-blue-800"
                onClick={() => {
                  const newWaitFor = [...(step.wait_for || []), ''];
                  onUpdate({ wait_for: newWaitFor });
                }}
              >
                + Add Wait For Step
              </button>
            </div>
          </div>
        </>
      )}

      {(step.type === 'fork' || step.type === 'parallel') && (
        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1">Parallel Steps</label>
          <div className="space-y-1">
            {step.parallel?.map((parallelStep, idx) => (
              <div key={idx} className="flex items-center gap-2">
                <select
                  value={parallelStep}
                  onChange={(e) => {
                    const newParallel = [...(step.parallel || [])];
                    newParallel[idx] = e.target.value;
                    onUpdate({ parallel: newParallel });
                  }}
                  className="flex-1 px-3 py-2 border border-slate-300 rounded-lg text-sm"
                >
                  <option value="">Select step</option>
                  {allSteps.filter(s => s !== step.name).map(s => (
                    <option key={s} value={s}>{s}</option>
                  ))}
                </select>
                <button
                  className="text-red-500 hover:text-red-700"
                  onClick={() => {
                    const newParallel = step.parallel?.filter((_, i) => i !== idx) || [];
                    onUpdate({ parallel: newParallel });
                  }}
                >
                  ×
                </button>
              </div>
            ))}
            <button
              className="text-sm text-blue-600 hover:text-blue-800"
              onClick={() => {
                const newParallel = [...(step.parallel || []), ''];
                onUpdate({ parallel: newParallel });
              }}
            >
              + Add Parallel Step
            </button>
          </div>
        </div>
      )}

      <div>
        <label className="block text-sm font-medium text-slate-700 mb-1">
          <input
            type="checkbox"
            checked={step.no_idempotent || false}
            onChange={(e) => onUpdate({ no_idempotent: e.target.checked })}
            className="mr-2"
          />
          No Idempotent
        </label>
      </div>
    </div>
  );
};

