import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { useS3Buckets, useS3CreateBucket, useS3DeleteBucket, useS3Objects, useS3DeleteObject } from '@/hooks/use-api'
import { Plus, Trash2, RefreshCw, Box, File, ArrowLeft } from 'lucide-react'

export function S3Explorer() {
  const [selectedBucket, setSelectedBucket] = useState<string | null>(null)
  const { data: buckets, isLoading, refetch } = useS3Buckets()
  const { data: objects } = useS3Objects(selectedBucket || '')
  const createBucket = useS3CreateBucket()
  const deleteBucket = useS3DeleteBucket()
  const deleteObject = useS3DeleteObject()

  const handleCreateBucket = async () => {
    const name = prompt('Enter bucket name:')
    if (name) {
      createBucket.mutate(name)
    }
  }

  const handleDeleteBucket = async (name: string) => {
    if (confirm(`Delete bucket "${name}"?`)) {
      deleteBucket.mutate(name)
    }
  }

  const handleDeleteObject = async (key: string) => {
    if (confirm(`Delete object "${key}"?`)) {
      deleteObject.mutate({ bucket: selectedBucket!, key })
    }
  }

  if (selectedBucket) {
    return (
      <div className="bg-slate-800/50 rounded-xl border border-slate-700 overflow-hidden">
        <div className="flex items-center justify-between px-4 py-3 border-b border-slate-700">
          <div className="flex items-center gap-3">
            <Button size="sm" variant="ghost" onClick={() => setSelectedBucket(null)} className="text-slate-400 hover:text-white">
              <ArrowLeft className="h-4 w-4 mr-1" />
              Back
            </Button>
            <Box className="h-5 w-5 text-orange-400" />
            <h2 className="font-semibold text-white">{selectedBucket}</h2>
            <span className="px-2 py-0.5 rounded text-xs bg-slate-700 text-slate-300">
              {objects?.length || 0} objects
            </span>
          </div>
        </div>
        <ScrollArea className="h-[500px]">
          {!objects || objects.length === 0 ? (
            <div className="p-4 text-slate-500 text-center">No objects in this bucket</div>
          ) : (
            <div className="divide-y divide-slate-700">
              {objects.map((obj: { key: string; size: number }) => (
                <div
                  key={obj.key}
                  className="flex items-center justify-between px-4 py-3 hover:bg-slate-700/30 group"
                >
                  <div className="flex items-center gap-3 min-w-0">
                    <File className="h-4 w-4 text-slate-400 flex-shrink-0" />
                    <span className="text-white font-mono text-sm truncate">{obj.key}</span>
                    <span className="text-slate-500 text-xs">({formatBytes(obj.size)})</span>
                  </div>
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => handleDeleteObject(obj.key)}
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

  return (
    <div className="bg-slate-800/50 rounded-xl border border-slate-700 overflow-hidden">
      <div className="flex items-center justify-between px-4 py-3 border-b border-slate-700">
        <div className="flex items-center gap-3">
          <Box className="h-5 w-5 text-orange-400" />
          <h2 className="font-semibold text-white">S3 Buckets</h2>
          <span className="px-2 py-0.5 rounded text-xs bg-slate-700 text-slate-300">
            {buckets?.length || 0} buckets
          </span>
        </div>
        <div className="flex gap-2">
          <Button size="sm" variant="outline" onClick={handleCreateBucket} className="bg-green-500/20 border-green-500/50 text-green-400 hover:bg-green-500/30">
            <Plus className="h-4 w-4 mr-1" />
            Create Bucket
          </Button>
          <Button size="sm" variant="ghost" onClick={() => refetch()} className="text-slate-400 hover:text-white">
            <RefreshCw className="h-4 w-4" />
          </Button>
        </div>
      </div>
      <ScrollArea className="h-[500px]">
        {isLoading ? (
          <div className="p-4 text-slate-400">Loading...</div>
        ) : !buckets || buckets.length === 0 ? (
          <div className="p-4 text-slate-500 text-center">No buckets. Create one to get started.</div>
        ) : (
          <div className="divide-y divide-slate-700">
            {buckets.map((bucket: { name: string }) => (
              <div
                key={bucket.name}
                className="flex items-center justify-between px-4 py-3 hover:bg-slate-700/30 cursor-pointer group"
                onClick={() => setSelectedBucket(bucket.name)}
              >
                <div className="flex items-center gap-3">
                  <Box className="h-5 w-5 text-orange-400" />
                  <span className="text-white font-medium">{bucket.name}</span>
                </div>
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={(e) => {
                    e.stopPropagation()
                    handleDeleteBucket(bucket.name)
                  }}
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

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}
