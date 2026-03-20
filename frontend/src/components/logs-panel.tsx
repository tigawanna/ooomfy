import { useState, useEffect, useRef } from 'react'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Copy, Trash2, Download, RefreshCw } from 'lucide-react'

export function LogsPanel() {
  const [logs, setLogs] = useState<string[]>([])
  const [isStreaming, setIsStreaming] = useState(true)
  const scrollRef = useRef<HTMLDivElement>(null)
  const eventSourceRef = useRef<EventSource | null>(null)

  useEffect(() => {
    async function fetchInitialLogs() {
      try {
        const res = await fetch('/api/logs/all')
        const data = await res.json()
        if (Array.isArray(data)) {
          setLogs(data)
        }
      } catch (err) {
        console.error('Failed to fetch logs:', err)
      }
    }

    fetchInitialLogs()

    const eventSource = new EventSource('/api/logs/stream')
    eventSourceRef.current = eventSource

    eventSource.onmessage = (event) => {
      const newLog = event.data
      if (newLog && newLog.trim()) {
        setLogs(prev => [...prev, newLog])
      }
    }

    eventSource.onerror = () => {
      setIsStreaming(false)
    }

    return () => {
      eventSource.close()
    }
  }, [])

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [logs])

  const copyLogs = () => {
    const text = logs.join('')
    navigator.clipboard.writeText(text)
  }

  const downloadLogs = () => {
    const text = logs.join('')
    const blob = new Blob([text], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `ooomfs-logs-${new Date().toISOString().split('T')[0]}.log`
    a.click()
    URL.revokeObjectURL(url)
  }

  const clearLogs = () => {
    setLogs([])
  }

  const reconnect = () => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
    }
    setIsStreaming(true)
    window.location.reload()
  }

  return (
    <div className="bg-slate-800/50 rounded-xl border border-slate-700 overflow-hidden">
      <div className="flex items-center justify-between px-4 py-3 border-b border-slate-700">
        <div className="flex items-center gap-2">
          <h2 className="font-semibold text-white">Logs</h2>
          <span className={`px-2 py-0.5 rounded text-xs ${isStreaming ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'}`}>
            {isStreaming ? 'Live' : 'Disconnected'}
          </span>
        </div>
        <div className="flex gap-2">
          <Button size="sm" variant="ghost" onClick={copyLogs} className="text-slate-400 hover:text-white">
            <Copy className="h-4 w-4 mr-1" />
            Copy
          </Button>
          <Button size="sm" variant="ghost" onClick={downloadLogs} className="text-slate-400 hover:text-white">
            <Download className="h-4 w-4 mr-1" />
            Export
          </Button>
          <Button size="sm" variant="ghost" onClick={clearLogs} className="text-slate-400 hover:text-white">
            <Trash2 className="h-4 w-4 mr-1" />
            Clear
          </Button>
          {!isStreaming && (
            <Button size="sm" variant="ghost" onClick={reconnect} className="text-yellow-400 hover:text-yellow-300">
              <RefreshCw className="h-4 w-4 mr-1" />
              Reconnect
            </Button>
          )}
        </div>
      </div>
      <ScrollArea className="h-[500px]" ref={scrollRef}>
        <div className="p-4 font-mono text-sm">
          {logs.length === 0 ? (
            <p className="text-slate-500">Waiting for logs...</p>
          ) : (
            logs.map((log, i) => (
              <div key={i} className="text-slate-300 whitespace-pre-wrap break-all leading-relaxed hover:bg-slate-700/30 -mx-2 px-2 py-0.5">
                {log}
              </div>
            ))
          )}
        </div>
      </ScrollArea>
    </div>
  )
}
