import React from 'react';

interface JsonViewerProps {
  data: any;
  className?: string;
  maxHeight?: string;
}

export const JsonViewer: React.FC<JsonViewerProps> = ({ data, className = '', maxHeight = '600px' }) => {
  const formatValue = (value: any, indent: number = 0): string => {
    if (value === null) return '<span class="json-null">null</span>';
    if (value === undefined) return '<span class="json-undefined">undefined</span>';
    
    if (typeof value === 'string') {
      return `<span class="json-string">${JSON.stringify(value)}</span>`;
    }
    
    if (typeof value === 'number') {
      return `<span class="json-number">${value}</span>`;
    }
    
    if (typeof value === 'boolean') {
      return `<span class="json-boolean">${value ? 'true' : 'false'}</span>`;
    }
    
    if (Array.isArray(value)) {
      if (value.length === 0) return '<span class="json-bracket">[]</span>';
      const items = value.map((item, idx) => 
        `${'  '.repeat(indent + 1)}${formatValue(item, indent + 1)}${idx < value.length - 1 ? ',' : ''}`
      ).join('\n');
      return `<span class="json-bracket">[</span>\n${items}\n${'  '.repeat(indent)}<span class="json-bracket">]</span>`;
    }
    
    if (typeof value === 'object') {
      const keys = Object.keys(value);
      if (keys.length === 0) return '<span class="json-bracket">{}</span>';
      const items = keys.map((key, idx) => 
        `${'  '.repeat(indent + 1)}<span class="json-key">"${key}"</span>: ${formatValue(value[key], indent + 1)}${idx < keys.length - 1 ? ',' : ''}`
      ).join('\n');
      return `<span class="json-bracket">{</span>\n${items}\n${'  '.repeat(indent)}<span class="json-bracket">}</span>`;
    }
    
    return String(value);
  };

  let formatted: string;
  try {
    const parsed = typeof data === 'string' ? JSON.parse(data) : data;
    formatted = JSON.stringify(parsed, null, 2);
  } catch {
    formatted = typeof data === 'string' ? data : JSON.stringify(data, null, 2);
  }

  // Apply syntax highlighting
  const highlighted = formatted
    .replace(/("(?:[^"\\]|\\.)*"):/g, '<span class="json-key">$1</span>:')
    .replace(/:\s*(true|false|null)/g, ': <span class="json-boolean">$1</span>')
    .replace(/:\s*(\d+\.?\d*)/g, ': <span class="json-number">$1</span>')
    .replace(/:\s*("(?:[^"\\]|\\.)*")/g, ': <span class="json-string">$1</span>')
    .replace(/(\{|\}|\[|\])/g, '<span class="json-bracket">$1</span>');

  return (
    <div className={`json-viewer ${className}`} style={{ maxHeight }}>
      <pre 
        className="whitespace-pre-wrap break-words m-0 p-0" 
        dangerouslySetInnerHTML={{ __html: highlighted }}
      />
    </div>
  );
};
