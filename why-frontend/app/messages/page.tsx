'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { isAuthenticated, logout } from '@/lib/auth'
import { listMessages } from '@/lib/api'
import { Message } from '@/lib/types'
import MessageForm from '@/components/MessageForm'
import MessageCard from '@/components/MessageCard'

export default function MessagesPage() {
  const router = useRouter()
  const [messages, setMessages] = useState<Message[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!isAuthenticated()) {
      router.push('/login')
      return
    }

    loadMessages()
  }, [router])

  const loadMessages = async () => {
    try {
      const data = await listMessages()
      setMessages(data || [])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load messages')
    } finally {
      setLoading(false)
    }
  }

  const handleLogout = () => {
    logout()
    router.push('/')
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <p className="text-gray-600">Loading...</p>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-white shadow-sm">
        <div className="max-w-3xl mx-auto px-4 py-4 flex items-center justify-between">
          <h1 className="text-2xl font-bold text-gray-900">Why</h1>
          <button
            onClick={handleLogout}
            className="text-sm text-gray-600 hover:text-gray-900"
          >
            Log out
          </button>
        </div>
      </nav>

      <main className="max-w-3xl mx-auto px-4 py-8">
        <MessageForm onSuccess={loadMessages} />

        {error && (
          <div className="mb-4 p-4 bg-red-50 text-red-800 rounded-md">
            {error}
          </div>
        )}

        {messages.length === 0 ? (
          <div className="text-center py-12 text-gray-500">
            No messages yet. Be the first to post!
          </div>
        ) : (
          <div>
            {messages.map((message) => (
              <MessageCard key={message.id} message={message} />
            ))}
          </div>
        )}
      </main>
    </div>
  )
}
