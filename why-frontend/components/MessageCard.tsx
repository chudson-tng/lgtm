import Link from 'next/link'
import { Message } from '@/lib/types'

interface MessageCardProps {
  message: Message
}

export default function MessageCard({ message }: MessageCardProps) {
  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    return date.toLocaleString()
  }

  return (
    <div className="bg-white rounded-lg shadow p-6 mb-4">
      <div className="flex items-start justify-between mb-2">
        <div className="text-sm text-gray-500">
          {formatDate(message.created_at)}
        </div>
      </div>
      <p className="text-gray-900 mb-4">{message.content}</p>
      {message.media_urls && message.media_urls.length > 0 && (
        <div className="grid grid-cols-2 gap-2 mb-4">
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
      <Link
        href={`/messages/${message.id}`}
        className="text-blue-600 hover:text-blue-700 text-sm font-medium"
      >
        View replies →
      </Link>
    </div>
  )
}
