'use client'

import { useState } from 'react'
import { createReply, uploadMedia } from '@/lib/api'

interface ReplyFormProps {
  messageId: string
  onSuccess?: () => void
}

export default function ReplyForm({ messageId, onSuccess }: ReplyFormProps) {
  const [content, setContent] = useState('')
  const [files, setFiles] = useState<File[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!content.trim()) return

    setLoading(true)
    setError('')

    try {
      // Upload files first
      const mediaUrls: string[] = []
      for (const file of files) {
        const url = await uploadMedia(file)
        mediaUrls.push(url)
      }

      // Create reply
      await createReply(messageId, content, mediaUrls)

      // Reset form
      setContent('')
      setFiles([])
      if (onSuccess) onSuccess()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to post reply')
    } finally {
      setLoading(false)
    }
  }

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      setFiles(Array.from(e.target.files))
    }
  }

  return (
    <form onSubmit={handleSubmit} className="bg-white rounded-lg shadow p-4 mb-4">
      {error && (
        <div className="mb-4 p-3 bg-red-50 text-red-800 rounded-md text-sm">
          {error}
        </div>
      )}
      <textarea
        value={content}
        onChange={(e) => setContent(e.target.value)}
        placeholder="Write a reply..."
        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
        rows={2}
        disabled={loading}
      />
      <div className="mt-2 flex items-center justify-between">
        <input
          type="file"
          accept="image/*,video/*"
          multiple
          onChange={handleFileChange}
          className="text-sm text-gray-600"
          disabled={loading}
        />
        <button
          type="submit"
          disabled={loading || !content.trim()}
          className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {loading ? 'Replying...' : 'Reply'}
        </button>
      </div>
      {files.length > 0 && (
        <div className="mt-2 text-sm text-gray-600">
          {files.length} file(s) selected
        </div>
      )}
    </form>
  )
}
