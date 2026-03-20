import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { useRedisKeys, useDeleteRedisKey } from '@/hooks/use-api'
import { Trash2, RefreshCw, Database, Search } from 'lucide-react'

export function RedisExplorer() {
  const [filter, setFilter] = useState('')
  const { data: keys, isLoading, refetch } = useRedisKeys()
  const deleteKey = useDeleteRedisKey()

  const filteredKeys = keys?.filter((k: { key: string }) => 
    k.key.toLowerCase().includes(filter.toLowerCase())
  ) || []

  const handleDelete = async (key: string) => {
    if (confirm(`Delete key "${key}"?`)) {
      deleteKey.mutate(key)
    }
  }

  return (
    <div className="bg-slate-800/50 rounded-xl border border-slate-700 overflow-hidden">
      <div className="flex items-center justify-between px-4 py-3 border-b border-slate-700">
        <div className="flex items-center gap-3">
          <Database className="h-5 w-5 text-red-400" />
          <h2 className="font-semibold text-white">Redis Explorer</h2>
          <span className="px-2 py-0.5 rounded text-xs bg-slate-700 text-slate-300">
            {filteredKeys.length} keys
          </span>
        </div>
        <div className="flex items-center gap-2">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
            <input
              type="text"
              placeholder="Filter keys..."
              value={filter}
              onChange={(e) => setFilter(e.target.value)}
              className="pl-9 pr-4 py-1.5 bg-slate-700 border border-slate-600 rounded-lg text-sm text-white placeholder-slate-400 focus:outline-none focus:border-slate-500 w-48"
            />
          </div>
          <Button size="sm" variant="ghost" onClick={() => refetch()} className="text-slate-400 hover:text-white">
            <RefreshCw className="h-4 w-4" />
          </Button>
        </div>
      </div>
      <ScrollArea className="h-[500px]">
        {isLoading ? (
          <div className="p-4 text-slate-400">Loading...</div>
        ) : filteredKeys.length === 0 ? (
          <div className="p-4 text-slate-500 text-center">
            {keys?.length === 0 ? 'No keys in Redis' : 'No keys match filter'}
          </div>
        ) : (
          <div className="divide-y divide-slate-700">
            {filteredKeys.map((item: { key: string; type: string }) => (
              <div
                key={item.key}
                className="flex items-center justify-between px-4 py-3 hover:bg-slate-700/30 group"
              >
                <div className="flex items-center gap-3 min-w-0">
                  <span className={`px-2 py-0.5 rounded text-xs font-medium ${
                    item.type === 'string' ? 'bg-blue-500/20 text-blue-400' :
                    item.type === 'list' ? 'bg-purple-500/20 text-purple-400' :
                    item.type === 'hash' ? 'bg-yellow-500/20 text-yellow-400' :
                    item.type === 'set' ? 'bg-green-500/20 text-green-400' :
                    'bg-slate-500/20 text-slate-400'
                  }`}>
                    {item.type}
                  </span>
                  <span className="text-white font-mono text-sm truncate">{item.key}</span>
                </div>
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={() => handleDelete(item.key)}
                  className="opacity-0 group-hover:opacity-100 text-red-400 hover:text-red-300 hover:bg-red-500/20"
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            ))}
          </div>
        )}
      </ScrollArea>
    </div>
  )
}
