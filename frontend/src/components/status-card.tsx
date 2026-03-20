import { LucideIcon } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useServiceControl } from '@/hooks/use-api'
import { Play, Square, RotateCcw } from 'lucide-react'

interface StatusCardProps {
  name: string
  icon: LucideIcon
  status: 'running' | 'stopped'
  port: number
  service?: string
  isDashboard?: boolean
}

export function StatusCard({ name, icon: Icon, status, port, service, isDashboard }: StatusCardProps) {
  const { start, stop, restart } = useServiceControl(service || '')
  const isRunning = status === 'running'

  return (
    <div className="bg-slate-800/50 rounded-xl border border-slate-700 p-6 hover:border-slate-600 transition-colors">
      <div className="flex items-start justify-between mb-4">
        <div className="flex items-center gap-3">
          <div className={`h-10 w-10 rounded-lg flex items-center justify-center ${isRunning ? 'bg-green-500/20 text-green-400' : 'bg-slate-700 text-slate-400'}`}>
            <Icon className="h-5 w-5" />
          </div>
          <div>
            <h3 className="font-semibold text-white">{name}</h3>
            <p className="text-sm text-slate-400">:{port}</p>
          </div>
        </div>
        <span className={`px-2.5 py-1 rounded-full text-xs font-medium ${isRunning ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'}`}>
          {isRunning ? 'Running' : 'Stopped'}
        </span>
      </div>
      
      {!isDashboard && (
        <div className="flex gap-2">
          {isRunning ? (
            <Button size="sm" variant="outline" onClick={() => stop.mutate()} className="flex-1 bg-slate-700/50 border-slate-600 hover:bg-red-500/20 hover:text-red-400 hover:border-red-500/50">
              <Square className="h-4 w-4 mr-1" />
              Stop
            </Button>
          ) : (
            <Button size="sm" variant="outline" onClick={() => start.mutate()} className="flex-1 bg-slate-700/50 border-slate-600 hover:bg-green-500/20 hover:text-green-400 hover:border-green-500/50">
              <Play className="h-4 w-4 mr-1" />
              Start
            </Button>
          )}
          <Button size="sm" variant="outline" onClick={() => restart.mutate()} className="bg-slate-700/50 border-slate-600 hover:bg-yellow-500/20 hover:text-yellow-400 hover:border-yellow-500/50">
            <RotateCcw className="h-4 w-4" />
          </Button>
        </div>
      )}
    </div>
  )
}
