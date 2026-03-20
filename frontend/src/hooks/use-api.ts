import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'

const API_BASE = ''

export interface Status {
  redis: 'running' | 'stopped'
  s3: 'running' | 'stopped'
  smtp: 'running' | 'stopped'
  dashboard: 'running' | 'stopped'
}

export function useStatus() {
  return useQuery<Status>({
    queryKey: ['status'],
    queryFn: async () => {
      const res = await fetch(`${API_BASE}/api/status`)
      return res.json()
    },
  })
}

export function useServiceControl(service: string) {
  const queryClient = useQueryClient()

  const start = useMutation({
    mutationFn: async () => {
      await fetch(`${API_BASE}/api/${service}/start`, { method: 'POST' })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['status'] })
    },
  })

  const stop = useMutation({
    mutationFn: async () => {
      await fetch(`${API_BASE}/api/${service}/stop`, { method: 'POST' })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['status'] })
    },
  })

  const restart = useMutation({
    mutationFn: async () => {
      await fetch(`${API_BASE}/api/${service}/restart`, { method: 'POST' })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['status'] })
    },
  })

  return { start, stop, restart }
}

export function useRedisKeys() {
  return useQuery({
    queryKey: ['redis', 'keys'],
    queryFn: async () => {
      const res = await fetch(`${API_BASE}/api/redis/keys`)
      return res.json()
    },
  })
}

export function useRedisKey(key: string) {
  return useQuery({
    queryKey: ['redis', 'key', key],
    queryFn: async () => {
      const res = await fetch(`${API_BASE}/api/redis/key/${encodeURIComponent(key)}`)
      return res.json()
    },
    enabled: !!key,
  })
}

export function useDeleteRedisKey() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (key: string) => {
      await fetch(`${API_BASE}/api/redis/key/${encodeURIComponent(key)}`, { method: 'DELETE' })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['redis', 'keys'] })
    },
  })
}

export function useS3Buckets() {
  return useQuery({
    queryKey: ['s3', 'buckets'],
    queryFn: async () => {
      const res = await fetch(`${API_BASE}/api/s3/buckets`)
      return res.json()
    },
  })
}

export function useS3Objects(bucket: string) {
  return useQuery({
    queryKey: ['s3', 'bucket', bucket, 'objects'],
    queryFn: async () => {
      const res = await fetch(`${API_BASE}/api/s3/bucket/${encodeURIComponent(bucket)}/objects`)
      return res.json()
    },
    enabled: !!bucket,
  })
}

export function useS3CreateBucket() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (bucket: string) => {
      await fetch(`${API_BASE}/api/s3/bucket/${encodeURIComponent(bucket)}`, { method: 'PUT' })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['s3', 'buckets'] })
    },
  })
}

export function useS3DeleteBucket() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (bucket: string) => {
      await fetch(`${API_BASE}/api/s3/bucket/${encodeURIComponent(bucket)}`, { method: 'DELETE' })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['s3', 'buckets'] })
    },
  })
}

export function useS3DeleteObject() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({ bucket, key }: { bucket: string; key: string }) => {
      await fetch(`${API_BASE}/api/s3/bucket/${encodeURIComponent(bucket)}/object/${encodeURIComponent(key)}`, { method: 'DELETE' })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['s3'] })
    },
  })
}

export function useSMTPEmails() {
  return useQuery({
    queryKey: ['smtp', 'emails'],
    queryFn: async () => {
      const res = await fetch(`${API_BASE}/api/smtp/emails`)
      return res.json()
    },
  })
}

export function useSMTPEmail(id: string) {
  return useQuery({
    queryKey: ['smtp', 'email', id],
    queryFn: async () => {
      const res = await fetch(`${API_BASE}/api/smtp/email/${encodeURIComponent(id)}`)
      return res.json()
    },
    enabled: !!id,
  })
}

export function useDeleteSMTPEmail() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => {
      await fetch(`${API_BASE}/api/smtp/email/${encodeURIComponent(id)}`, { method: 'DELETE' })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['smtp', 'emails'] })
    },
  })
}

export function useClearSMTPEmails() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async () => {
      await fetch(`${API_BASE}/api/smtp/clear`, { method: 'POST' })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['smtp', 'emails'] })
    },
  })
}

export function useRedisStats() {
  return useQuery({
    queryKey: ['redis', 'stats'],
    queryFn: async () => {
      const res = await fetch(`${API_BASE}/api/redis/`)
      return res.json()
    },
  })
}

export function useS3Stats() {
  return useQuery({
    queryKey: ['s3', 'stats'],
    queryFn: async () => {
      const res = await fetch(`${API_BASE}/api/s3/`)
      return res.json()
    },
  })
}

export function useSMTPStats() {
  return useQuery({
    queryKey: ['smtp', 'stats'],
    queryFn: async () => {
      const res = await fetch(`${API_BASE}/api/smtp/`)
      return res.json()
    },
  })
}

export function usePersist() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async () => {
      await fetch(`${API_BASE}/api/persist`, { method: 'POST' })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['status'] })
    },
  })
}

export function useRedisSave() {
  return useMutation({
    mutationFn: async () => {
      await fetch(`${API_BASE}/api/redis/save`, { method: 'POST' })
    },
  })
}
