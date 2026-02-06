'use client'

import { useEffect, useState } from 'react'
import { useRouter, useParams } from 'next/navigation'
import Link from 'next/link'
import { isAuthenticated } from '@/lib/auth'
import { getMessage, listReplies } from '@/lib/api'
import { Message, Reply } from '@/lib/types'
import ReplyForm from '@/components/ReplyForm'

export default function MessageDetailPage() {
  const router = useRouter()
  const params = useParams()
  const messageId = params.id as string

  const [message, setMessage] = useState<Message | null>(null)
  const [replies, setReplies] = useState<Reply[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!isAuthenticated()) {
      router.push('/login')
      return
    }

    loadMessageAndReplies()
  }, [messageId, router])

  const loadMessageAndReplies = async () => {
    try {
      const [messageData, repliesData] = await Promise.all([
        getMessage(messageId),
        listReplies(messageId),
      ])
      setMessage(messageData)
      setReplies(repliesData || [])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load message')
    } finally {
      setLoading(false)
    }
  }

  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    return date.toLocaleString()
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <p className="text-gray-600">Loading...</p>
      </div>
    )
  }

  if (error || !message) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <p className="text-red-600 mb-4">{error || 'Message not found'}</p>
          <Link href="/messages" className="text-blue-600 hover:text-blue-700">
            ← Back to messages
          </Link>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-white shadow-sm">
        <div className="max-w-3xl mx-auto px-4 py-4">
          <Link
            href="/messages"
            className="text-blue-600 hover:text-blue-700 text-sm font-medium"
          >
            ← Back to messages
          </Link>
        </div>
      </nav>

      <main className="max-w-3xl mx-auto px-4 py-8">
        {/* Original message */}
        <div className="bg-white rounded-lg shadow p-6 mb-8">
          <div className="text-sm text-gray-500 mb-2">
            {formatDate(message.created_at)}
          </div>
          <p className="text-gray-900 text-lg mb-4">{message.content}</p>
          {message.media_urls && message.media_urls.length > 0 && (
            <div className="grid grid-cols-2 gap-2">
              {message.media_urls.map((url, index) => (
                <div key={index} className="relative aspect-video bg-gray-100 rounded">
                  {url.match(/\.(jpg|jpeg|png|gif)$/i) ? (
                    <img
                      src={url}
                      alt={`Media ${index + 1}`}
                      className="w-full h-full object-cover rounded"
                    />
                  ) : (
                    <video
                      src={url}
                      controls
                      className="w-full h-full object-cover rounded"
                    />
                  )}
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Reply form */}
        <div className="mb-8">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Replies</h2>
          <ReplyForm messageId={messageId} onSuccess={loadMessageAndReplies} />
        </div>

        {/* Replies list */}
        {replies.length === 0 ? (
          <div className="text-center py-8 text-gray-500">
            No replies yet. Be the first to reply!
          </div>
        ) : (
          <div className="space-y-4">
            {replies.map((reply) => (
              <div key={reply.id} className="bg-white rounded-lg shadow p-4">
                <div className="text-sm text-gray-500 mb-2">
                  {formatDate(reply.created_at)}
                </div>
                <p className="text-gray-900">{reply.content}</p>
                {reply.media_urls && reply.media_urls.length > 0 && (
                  <div className="grid grid-cols-2 gap-2 mt-3">
                    {reply.media_urls.map((url, index) => (
                      <div key={index} className="relative aspect-video bg-gray-100 rounded">
                        {url.match(/\.(jpg|jpeg|png|gif)$/i) ? (
                          <img
                            src={url}
                            alt={`Media ${index + 1}`}
                            className="w-full h-full object-cover rounded"
                          />
                        ) : (
                          <video
                            src={url}
                            controls
                            className="w-full h-full object-cover rounded"
                          />
                        )}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </main>
    </div>
  )
}
