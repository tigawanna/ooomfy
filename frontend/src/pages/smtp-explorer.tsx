import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { useSMTPEmails, useDeleteSMTPEmail, useClearSMTPEmails, useSMTPEmail } from '@/hooks/use-api'
import { Trash2, RefreshCw, Mail, X, ChevronRight } from 'lucide-react'

export function SMTPExplorer() {
  const [selectedEmail, setSelectedEmail] = useState<string | null>(null)
  const { data: emails, isLoading, refetch } = useSMTPEmails()
  const { data: emailDetails } = useSMTPEmail(selectedEmail || '')
  const deleteEmail = useDeleteSMTPEmail()
  const clearEmails = useClearSMTPEmails()

  const handleDelete = async (id: string) => {
    if (confirm('Delete this email?')) {
      deleteEmail.mutate(id)
      if (selectedEmail === id) {
        setSelectedEmail(null)
      }
    }
  }

  const handleClearAll = () => {
    if (confirm('Delete all emails?')) {
      clearEmails.mutate()
      setSelectedEmail(null)
    }
  }

  if (selectedEmail && emailDetails) {
    return (
      <div className="bg-slate-800/50 rounded-xl border border-slate-700 overflow-hidden">
        <div className="flex items-center justify-between px-4 py-3 border-b border-slate-700">
          <Button size="sm" variant="ghost" onClick={() => setSelectedEmail(null)} className="text-slate-400 hover:text-white">
            <X className="h-4 w-4 mr-1" />
            Close
          </Button>
          <Button
            size="sm"
            variant="ghost"
            onClick={() => handleDelete(selectedEmail)}
            className="text-red-400 hover:text-red-300 hover:bg-red-500/20"
          >
            <Trash2 className="h-4 w-4 mr-1" />
            Delete
          </Button>
        </div>
        <ScrollArea className="h-[500px]">
          <div className="p-6 space-y-4">
            <div>
              <label className="text-xs text-slate-400 uppercase tracking-wider">From</label>
              <p className="text-white">{emailDetails.from}</p>
            </div>
            <div>
              <label className="text-xs text-slate-400 uppercase tracking-wider">To</label>
              <p className="text-white">{emailDetails.to?.join(', ')}</p>
            </div>
            <div>
              <label className="text-xs text-slate-400 uppercase tracking-wider">Subject</label>
              <p className="text-white font-medium">{emailDetails.subject}</p>
            </div>
            <div>
              <label className="text-xs text-slate-400 uppercase tracking-wider">Date</label>
              <p className="text-slate-300">{new Date(emailDetails.date).toLocaleString()}</p>
            </div>
            <div className="pt-4 border-t border-slate-700">
              <label className="text-xs text-slate-400 uppercase tracking-wider mb-2 block">Body</label>
              <pre className="text-slate-300 whitespace-pre-wrap text-sm bg-slate-900/50 rounded-lg p-4 overflow-x-auto">
                {emailDetails.body || 'No body'}
              </pre>
            </div>
          </div>
        </ScrollArea>
      </div>
    )
  }

  return (
    <div className="bg-slate-800/50 rounded-xl border border-slate-700 overflow-hidden">
      <div className="flex items-center justify-between px-4 py-3 border-b border-slate-700">
        <div className="flex items-center gap-3">
          <Mail className="h-5 w-5 text-blue-400" />
          <h2 className="font-semibold text-white">Captured Emails</h2>
          <span className="px-2 py-0.5 rounded text-xs bg-slate-700 text-slate-300">
            {emails?.length || 0} emails
          </span>
        </div>
        <div className="flex gap-2">
          {emails && emails.length > 0 && (
            <Button size="sm" variant="outline" onClick={handleClearAll} className="bg-red-500/20 border-red-500/50 text-red-400 hover:bg-red-500/30">
              <Trash2 className="h-4 w-4 mr-1" />
              Clear All
            </Button>
          )}
          <Button size="sm" variant="ghost" onClick={() => refetch()} className="text-slate-400 hover:text-white">
            <RefreshCw className="h-4 w-4" />
          </Button>
        </div>
      </div>
      <ScrollArea className="h-[500px]">
        {isLoading ? (
          <div className="p-4 text-slate-400">Loading...</div>
        ) : !emails || emails.length === 0 ? (
          <div className="p-4 text-slate-500 text-center">
            <Mail className="h-12 w-12 mx-auto mb-3 opacity-50" />
            <p>No emails captured yet.</p>
            <p className="text-sm mt-1">Send an email to localhost:1025 to capture it.</p>
          </div>
        ) : (
          <div className="divide-y divide-slate-700">
            {[...emails].reverse().map((email: { id: string; from: string; to: string[]; subject: string; date: string }) => (
              <div
                key={email.id}
                className="flex items-center justify-between px-4 py-3 hover:bg-slate-700/30 cursor-pointer group"
                onClick={() => setSelectedEmail(email.id)}
              >
                <div className="flex items-center gap-3 min-w-0 flex-1">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="text-slate-400 text-xs">#{email.id}</span>
                      <span className="text-white font-medium truncate">{email.subject || '(no subject)'}</span>
                    </div>
                    <p className="text-slate-400 text-sm truncate">From: {email.from}</p>
                    <p className="text-slate-500 text-xs">{new Date(email.date).toLocaleString()}</p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={(e) => {
                      e.stopPropagation()
                      handleDelete(email.id)
                    }}
                    className="opacity-0 group-hover:opacity-100 text-red-400 hover:text-red-300 hover:bg-red-500/20"
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                  <ChevronRight className="h-4 w-4 text-slate-500" />
                </div>
              </div>
            ))}
          </div>
        )}
      </ScrollArea>
    </div>
  )
}
