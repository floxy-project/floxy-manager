import React from 'react';

export interface StepDefinition {
    name: string;
    type: 'task' | 'parallel' | 'condition' | 'fork' | 'join' | 'save_point' | 'human';
    handler?: string;
    max_retries?: number;
    next?: string[];
    prev?: string;
    else?: string;
    on_failure?: string;
    condition?: string;
    parallel?: string[];
    wait_for?: string[];
    join_strategy?: 'all' | 'any';
    metadata?: Record<string, any>;
    no_idempotent?: boolean;
}

export interface GraphDefinition {
    steps: Record<string, StepDefinition>;
    start: string;
}

export interface WorkflowStep {
    id: number;
    instance_id: number;
    step_name: string;
    step_type: string;
    status: 'pending' | 'running' | 'completed' | 'failed' | 'skipped' | 'compensation' | 'rolled_back' | 'waiting_decision' | 'confirmed' | 'rejected' | 'paused';
    input?: any;
    output?: any;
    error?: string;
    retry_count: number;
    max_retries: number;
    compensation_retry_count: number;
    started_at?: string;
    completed_at?: string;
    created_at: string;
}

interface WorkflowGraphProps {
    definition: GraphDefinition;
    steps?: WorkflowStep[];
    title?: string;
    showLegend?: boolean;
}

interface GraphNode {
    id: string;
    step: StepDefinition;
    x: number;
    y: number;
    status?: string;
    width: number;
    height: number;
}

interface GraphEdge {
    from: string;
    to: string;
    type: 'normal' | 'failure' | 'else' | 'parallel';
}

export const WorkflowGraph: React.FC<WorkflowGraphProps> = ({
                                                                definition,
                                                                steps = [],
                                                                title = "Workflow Graph",
                                                                showLegend = true
                                                            }) => {
    const [nodes, setNodes] = React.useState<GraphNode[]>([]);
    const [edges, setEdges] = React.useState<GraphEdge[]>([]);
    const [dimensions, setDimensions] = React.useState({ width: 1000, height: 700 });

    const stepStatusMap = React.useMemo(() => {
        const map = new Map<string, string>();
        steps.forEach(step => {
            map.set(step.step_name, step.status);
        });
        return map;
    }, [steps]);

    const getStepStatus = (stepName: string): string => {
        return stepStatusMap.get(stepName) || 'pending';
    };

    const getStepGradient = (stepName: string): string => {
        const status = getStepStatus(stepName);
        switch (status) {
            case 'completed': return 'linear-gradient(135deg, #10b981 0%, #059669 100%)';
            case 'failed': return 'linear-gradient(135deg, #ef4444 0%, #dc2626 100%)';
            case 'running': return 'linear-gradient(135deg, #3b82f6 0%, #2563eb 100%)';
            case 'compensation': return 'linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%)';
            case 'rolled_back': return 'linear-gradient(135deg, #f59e0b 0%, #d97706 100%)';
            case 'skipped': return 'linear-gradient(135deg, #6b7280 0%, #4b5563 100%)';
            case 'waiting_decision': return 'linear-gradient(135deg, #f59e0b 0%, #d97706 100%)';
            case 'confirmed': return 'linear-gradient(135deg, #10b981 0%, #059669 100%)';
            case 'rejected': return 'linear-gradient(135deg, #ef4444 0%, #dc2626 100%)';
            case 'paused': return 'linear-gradient(135deg, #6366f1 0%, #4f46e5 100%)';
            default: return 'linear-gradient(135deg, #6366f1 0%, #4f46e5 100%)';
        }
    };

    const getStepColor = (type: string): string => {
        switch (type) {
            case 'human': return 'linear-gradient(135deg, #f59e0b 0%, #d97706 100%)';
            case 'task': return 'linear-gradient(135deg, #6366f1 0%, #4f46e5 100%)';
            case 'parallel': return 'linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%)';
            case 'condition': return 'linear-gradient(135deg, #06b6d4 0%, #0891b2 100%)';
            case 'fork': return 'linear-gradient(135deg, #ec4899 0%, #db2777 100%)';
            case 'join': return 'linear-gradient(135deg, #10b981 0%, #059669 100%)';
            case 'save_point': return 'linear-gradient(135deg, #6b7280 0%, #4b5563 100%)';
            default: return 'linear-gradient(135deg, #6366f1 0%, #4f46e5 100%)';
        }
    };

    const getStepIcon = (type: string): string => {
        switch (type) {
            case 'task': return '⚙️';
            case 'parallel': return '∥';
            case 'condition': return '❓';
            case 'fork': return '🔀';
            case 'join': return '🔗';
            case 'save_point': return '💾';
            case 'human': return '👤';
            default: return '📋';
        }
    };

    const calculateLayout = React.useCallback(() => {
        const nodeMap = new Map<string, GraphNode>();
        const edgeList: GraphEdge[] = [];
        const visited = new Set<string>();
        const levels: string[][] = [];

        const traverse = (stepName: string, level: number) => {
            if (visited.has(stepName) || !definition.steps[stepName]) return;

            visited.add(stepName);
            const step = definition.steps[stepName];

            if (!levels[level]) levels[level] = [];
            levels[level].push(stepName);

            const node: GraphNode = {
                id: stepName,
                step,
                x: 0,
                y: 0,
                status: getStepStatus(stepName),
                width: 120,
                height: 60
            };
            nodeMap.set(stepName, node);

            // Handle parallel steps
            if (step.parallel) {
                step.parallel.forEach(parallelStep => {
                    edgeList.push({ from: stepName, to: parallelStep, type: 'parallel' });
                    traverse(parallelStep, level + 1);
                });

                // For parallel steps, connect them to the next step after parallel
                if (step.next) {
                    step.next.forEach(nextStep => {
                        step.parallel!.forEach(parallelStep => {
                            edgeList.push({ from: parallelStep, to: nextStep, type: 'normal' });
                        });
                        traverse(nextStep, level + 2);
                    });
                }
            } else {
                // Handle normal next steps
                if (step.next) {
                    step.next.forEach(nextStep => {
                        edgeList.push({ from: stepName, to: nextStep, type: 'normal' });
                        traverse(nextStep, level + 1);
                    });
                }
            }

            if (step.else) {
                edgeList.push({ from: stepName, to: step.else, type: 'else' });
                traverse(step.else, level + 1);
            }

            // Handle fork+join pattern
            if (step.type === 'fork') {
                // Fork connects to first steps of each branch
                if (step.next) {
                    step.next.forEach(branchStart => {
                        edgeList.push({ from: stepName, to: branchStart, type: 'normal' });
                        traverse(branchStart, level + 1);
                    });
                }
                // Don't traverse next steps from fork - they will be handled by join
                return;
            }

            // Handle join - find all steps that should connect to join
            if (step.type === 'join') {
                // Find all steps that should connect to this join
                Object.keys(definition.steps).forEach(stepKey => {
                    const otherStep = definition.steps[stepKey];
                    if (otherStep.next && otherStep.next.includes(stepName)) {
                        // This step should connect to join
                        edgeList.push({ from: stepKey, to: stepName, type: 'normal' });
                    }
                });
                // Continue traversal from join
                if (step.next) {
                    step.next.forEach(nextStep => {
                        traverse(nextStep, level + 1);
                    });
                }
            }
        };

        traverse(definition.start, 0);

        const maxLevel = levels.length;
        const levelWidth = Math.max(200, dimensions.width / maxLevel);
        const nodeSpacing = 100;
        const padding = 50;

        levels.forEach((levelNodes, levelIndex) => {
            const startY = Math.max(padding, (dimensions.height - (levelNodes.length - 1) * nodeSpacing) / 2);

            levelNodes.forEach((nodeName, nodeIndex) => {
                const node = nodeMap.get(nodeName);
                if (node) {
                    node.x = levelIndex * levelWidth + padding;
                    node.y = startY + nodeIndex * nodeSpacing;
                }
            });
        });

        // Adjust dimensions based on actual content
        const maxX = Math.max(...Array.from(nodeMap.values()).map(n => n.x + n.width)) + padding;
        const maxY = Math.max(...Array.from(nodeMap.values()).map(n => n.y + n.height)) + padding;

        setDimensions(prev => ({
            width: Math.max(prev.width, maxX),
            height: Math.max(prev.height, maxY)
        }));

        // Filter out direct connections from fork to join
        const filteredEdges = edgeList.filter(edge => {
            const fromStep = definition.steps[edge.from];
            const toStep = definition.steps[edge.to];

            // Remove direct connection from fork to join
            if (fromStep && toStep && fromStep.type === 'fork' && toStep.type === 'join') {
                return false;
            }

            return true;
        });

        setNodes(Array.from(nodeMap.values()));
        setEdges(filteredEdges);
    }, [definition, dimensions, getStepStatus]);

    React.useEffect(() => {
        calculateLayout();
    }, [calculateLayout]);

    const getEdgeColor = (type: string): string => {
        switch (type) {
            case 'failure': return '#ef4444';
            case 'else': return '#f59e0b';
            case 'parallel': return '#8b5cf6';
            default: return '#6b7280';
        }
    };

    const getEdgeStyle = (type: string): string => {
        switch (type) {
            case 'failure': return 'dashed';
            case 'else': return 'dotted';
            case 'parallel': return 'solid';
            default: return 'solid';
        }
    };

    return (
        <div className="workflow-graph">
            <h3>{title}</h3>
            <div
                className="graph-container"
                style={{
                    width: '100%',
                    height: '600px',
                    position: 'relative',
                    border: '1px solid #e5e7eb',
                    borderRadius: '8px',
                    backgroundColor: '#f9fafb',
                    overflow: 'auto'
                }}
            >
                <svg
                    width={dimensions.width}
                    height={dimensions.height}
                    style={{ position: 'absolute', top: 0, left: 0 }}
                >
                    {edges.map((edge, index) => {
                        const fromNode = nodes.find(n => n.id === edge.from);
                        const toNode = nodes.find(n => n.id === edge.to);

                        if (!fromNode || !toNode) return null;

                        const startX = fromNode.x + fromNode.width / 2;
                        const startY = fromNode.y + fromNode.height / 2;
                        const endX = toNode.x + toNode.width / 2;
                        const endY = toNode.y + toNode.height / 2;

                        return (
                            <line
                                key={index}
                                x1={startX}
                                y1={startY}
                                x2={endX}
                                y2={endY}
                                stroke={getEdgeColor(edge.type)}
                                strokeWidth="2"
                                strokeDasharray={getEdgeStyle(edge.type) === 'dashed' ? '5,5' :
                                    getEdgeStyle(edge.type) === 'dotted' ? '2,3' : 'none'}
                                markerEnd="url(#arrowhead)"
                            />
                        );
                    })}

                    <defs>
                        <marker
                            id="arrowhead"
                            markerWidth="10"
                            markerHeight="7"
                            refX="9"
                            refY="3.5"
                            orient="auto"
                        >
                            <polygon
                                points="0 0, 10 3.5, 0 7"
                                fill="#6b7280"
                            />
                        </marker>
                    </defs>
                </svg>

                {nodes.map((node) => (
                    <div
                        key={node.id}
                        className="graph-node"
                        style={{
                            position: 'absolute',
                            left: node.x,
                            top: node.y,
                            width: node.width,
                            height: node.height,
                            background: node.step.type === 'human' ? getStepColor(node.step.type) : getStepGradient(node.id),
                            border: '1px solid rgba(255,255,255,0.2)',
                            borderRadius: '12px',
                            display: 'flex',
                            flexDirection: 'column',
                            alignItems: 'center',
                            justifyContent: 'center',
                            fontSize: '12px',
                            fontWeight: '600',
                            color: 'white',
                            cursor: 'pointer',
                            transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                            backdropFilter: 'blur(10px)',
                            boxShadow: '0 8px 32px rgba(0,0,0,0.1)'
                        }}
                        onMouseEnter={(e) => {
                            e.currentTarget.style.transform = 'scale(1.08) translateY(-2px)';
                            e.currentTarget.style.zIndex = '10';
                            e.currentTarget.style.boxShadow = '0 20px 40px rgba(0,0,0,0.2)';
                        }}
                        onMouseLeave={(e) => {
                            e.currentTarget.style.transform = 'scale(1) translateY(0)';
                            e.currentTarget.style.zIndex = '1';
                            e.currentTarget.style.boxShadow = '0 8px 32px rgba(0,0,0,0.1)';
                        }}
                    >
                        <div style={{ fontSize: '16px', marginBottom: '2px' }}>
                            {getStepIcon(node.step.type)}
                        </div>
                        <div style={{ textAlign: 'center', lineHeight: '1.2' }}>
                            <div style={{ fontWeight: '600' }}>{node.id}</div>
                            <div style={{ fontSize: '10px', opacity: 0.8 }}>
                                {node.step.type}
                            </div>
                            {node.step.handler && (
                                <div style={{ fontSize: '9px', opacity: 0.7 }}>
                                    {node.step.handler}
                                </div>
                            )}
                        </div>

                        {/* Compensation chip */}
                        {node.step.on_failure && (
                            <div
                                style={{
                                    position: 'absolute',
                                    top: '-10px',
                                    right: '-10px',
                                    background: 'linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%)',
                                    color: 'white',
                                    fontSize: '10px',
                                    fontWeight: '700',
                                    padding: '4px 8px',
                                    borderRadius: '12px',
                                    border: '2px solid rgba(255,255,255,0.3)',
                                    boxShadow: '0 4px 12px rgba(139, 92, 246, 0.4)',
                                    zIndex: 5,
                                    backdropFilter: 'blur(10px)',
                                    transition: 'all 0.3s ease',
                                    cursor: 'pointer'
                                }}
                                title={`Compensation: ${node.step.on_failure}`}
                                onMouseEnter={(e) => {
                                    e.currentTarget.style.transform = 'scale(1.1)';
                                    e.currentTarget.style.boxShadow = '0 6px 16px rgba(139, 92, 246, 0.6)';
                                }}
                                onMouseLeave={(e) => {
                                    e.currentTarget.style.transform = 'scale(1)';
                                    e.currentTarget.style.boxShadow = '0 4px 12px rgba(139, 92, 246, 0.4)';
                                }}
                            >
                                ⚡
                            </div>
                        )}
                    </div>
                ))}
            </div>

            {showLegend && (
                <div className="graph-legend">
                    <div>
                        <div style={{
                            width: '16px',
                            height: '16px',
                            background: 'linear-gradient(135deg, #10b981 0%, #059669 100%)',
                            borderRadius: '8px',
                            boxShadow: '0 2px 8px rgba(16, 185, 129, 0.3)'
                        }}></div>
                        <span>Completed</span>
                    </div>
                    <div>
                        <div style={{
                            width: '16px',
                            height: '16px',
                            background: 'linear-gradient(135deg, #ef4444 0%, #dc2626 100%)',
                            borderRadius: '8px',
                            boxShadow: '0 2px 8px rgba(239, 68, 68, 0.3)'
                        }}></div>
                        <span>Failed</span>
                    </div>
                    <div>
                        <div style={{
                            width: '16px',
                            height: '16px',
                            background: 'linear-gradient(135deg, #3b82f6 0%, #2563eb 100%)',
                            borderRadius: '8px',
                            boxShadow: '0 2px 8px rgba(59, 130, 246, 0.3)'
                        }}></div>
                        <span>Running</span>
                    </div>
                    <div>
                        <div style={{
                            width: '16px',
                            height: '16px',
                            background: 'linear-gradient(135deg, #f59e0b 0%, #d97706 100%)',
                            borderRadius: '8px',
                            boxShadow: '0 2px 8px rgba(245, 158, 11, 0.3)'
                        }}></div>
                        <span>Rolled Back</span>
                    </div>
                    <div>
                        <div style={{
                            width: '16px',
                            height: '16px',
                            background: 'linear-gradient(135deg, #6366f1 0%, #4f46e5 100%)',
                            borderRadius: '8px',
                            boxShadow: '0 2px 8px rgba(99, 102, 241, 0.3)'
                        }}></div>
                        <span>Pending</span>
                    </div>
                    <div>
                        <div style={{
                            width: '20px',
                            height: '20px',
                            background: 'linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%)',
                            borderRadius: '50%',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            fontSize: '10px',
                            color: 'white',
                            fontWeight: '700',
                            boxShadow: '0 4px 12px rgba(139, 92, 246, 0.4)'
                        }}>
                            ⚡
                        </div>
                        <span>Has Compensation</span>
                    </div>
                    <div>
                        <div style={{
                            width: '20px',
                            height: '20px',
                            background: 'linear-gradient(135deg, #f59e0b 0%, #d97706 100%)',
                            borderRadius: '50%',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            fontSize: '12px',
                            color: 'white',
                            fontWeight: '700',
                            boxShadow: '0 4px 12px rgba(245, 158, 11, 0.4)'
                        }}>
                            👤
                        </div>
                        <span>Human Step</span>
                    </div>
                    <div>
                        <div style={{
                            width: '16px',
                            height: '16px',
                            background: 'linear-gradient(135deg, #f59e0b 0%, #d97706 100%)',
                            borderRadius: '8px',
                            boxShadow: '0 2px 8px rgba(245, 158, 11, 0.3)'
                        }}></div>
                        <span>Waiting Decision</span>
                    </div>
                    <div>
                        <div style={{
                            width: '16px',
                            height: '16px',
                            background: 'linear-gradient(135deg, #10b981 0%, #059669 100%)',
                            borderRadius: '8px',
                            boxShadow: '0 2px 8px rgba(16, 185, 129, 0.3)'
                        }}></div>
                        <span>Confirmed</span>
                    </div>
                    <div>
                        <div style={{
                            width: '16px',
                            height: '16px',
                            background: 'linear-gradient(135deg, #ef4444 0%, #dc2626 100%)',
                            borderRadius: '8px',
                            boxShadow: '0 2px 8px rgba(239, 68, 68, 0.3)'
                        }}></div>
                        <span>Rejected</span>
                    </div>
                    <div>
                        <div style={{
                            width: '16px',
                            height: '16px',
                            background: 'linear-gradient(135deg, #6366f1 0%, #4f46e5 100%)',
                            borderRadius: '8px',
                            boxShadow: '0 2px 8px rgba(99, 102, 241, 0.3)'
                        }}></div>
                        <span>Paused</span>
                    </div>
                </div>
            )}
        </div>
    );
};
