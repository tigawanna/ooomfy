import { useState, useEffect } from 'react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Button } from '@/components/ui/button'
import { StatusCard } from '@/components/status-card'
import { LogsPanel } from '@/components/logs-panel'
import { RedisExplorer } from '@/pages/redis-explorer'
import { S3Explorer } from '@/pages/s3-explorer'
import { SMTPExplorer } from '@/pages/smtp-explorer'
import { useStatus } from '@/hooks/use-status'
import { Activity, Database, Mail, Server, Box, Terminal } from 'lucide-react'

function App() {
  const [activeTab, setActiveTab] = useState('dashboard')
  const { data: status, refetch } = useStatus()

  useEffect(() => {
    const interval = setInterval(refetch, 3000)
    return () => clearInterval(interval)
  }, [refetch])

  const tabs = [
    { value: 'dashboard', label: 'Dashboard', icon: Activity },
    { value: 'redis', label: 'Redis', icon: Database },
    { value: 's3', label: 'S3', icon: Box },
    { value: 'smtp', label: 'SMTP', icon: Mail },
    { value: 'logs', label: 'Logs', icon: Terminal },
  ]

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950">
      <header className="border-b border-slate-800 bg-slate-900/50 backdrop-blur-sm sticky top-0 z-50">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="h-10 w-10 rounded-lg bg-gradient-to-br from-cyan-500 to-blue-600 flex items-center justify-center">
                <Server className="h-6 w-6 text-white" />
              </div>
              <div>
                <h1 className="text-xl font-bold text-white">DevStack Manager</h1>
                <p className="text-xs text-slate-400">Local Development Environment</p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <span className="text-xs text-slate-500">Dashboard: :8080</span>
            </div>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-6">
        <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
          <TabsList className="bg-slate-800/50 border-slate-700">
            {tabs.map((tab) => (
              <TabsTrigger
                key={tab.value}
                value={tab.value}
                className="data-[state=active]:bg-slate-700"
              >
                <tab.icon className="h-4 w-4 mr-2" />
                {tab.label}
              </TabsTrigger>
            ))}
          </TabsList>

          <TabsContent value="dashboard" className="space-y-6">
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
              <StatusCard
                name="Redis"
                icon={Database}
                status={status?.redis || 'stopped'}
                port={6379}
                service="redis"
              />
              <StatusCard
                name="S3"
                icon={Box}
                status={status?.s3 || 'stopped'}
                port={9000}
                service="s3"
              />
              <StatusCard
                name="SMTP"
                icon={Mail}
                status={status?.smtp || 'stopped'}
                port={1025}
                service="smtp"
              />
              <StatusCard
                name="Dashboard"
                icon={Activity}
                status={status?.dashboard || 'stopped'}
                port={8080}
                isDashboard
              />
            </div>

            <div className="bg-slate-800/30 rounded-xl border border-slate-700 p-6">
              <h2 className="text-lg font-semibold text-white mb-4">Quick Actions</h2>
              <div className="flex flex-wrap gap-3">
                <Button
                  variant="outline"
                  onClick={() => setActiveTab('redis')}
                  className="bg-slate-700/50 border-slate-600 hover:bg-slate-600"
                >
                  Explore Redis
                </Button>
                <Button
                  variant="outline"
                  onClick={() => setActiveTab('s3')}
                  className="bg-slate-700/50 border-slate-600 hover:bg-slate-600"
                >
                  Explore S3
                </Button>
                <Button
                  variant="outline"
                  onClick={() => setActiveTab('smtp')}
                  className="bg-slate-700/50 border-slate-600 hover:bg-slate-600"
                >
                  View Emails
                </Button>
                <Button
                  variant="outline"
                  onClick={() => setActiveTab('logs')}
                  className="bg-slate-700/50 border-slate-600 hover:bg-slate-600"
                >
                  View Logs
                </Button>
              </div>
            </div>
          </TabsContent>

          <TabsContent value="redis">
            <RedisExplorer />
          </TabsContent>

          <TabsContent value="s3">
            <S3Explorer />
          </TabsContent>

          <TabsContent value="smtp">
            <SMTPExplorer />
          </TabsContent>

          <TabsContent value="logs">
            <LogsPanel />
          </TabsContent>
        </Tabs>
      </main>
    </div>
  )
}

export default App
