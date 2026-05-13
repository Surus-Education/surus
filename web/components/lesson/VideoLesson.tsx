"use client";

import { useEffect, useRef, useState, type ComponentType } from "react";
import dynamic from "next/dynamic";
import type { VideoDetail } from "@/lib/types";
import { ExternalLink } from "lucide-react";
import { TiptapRenderer } from "@/components/editor/TiptapRenderer";

const ReactPlayer = dynamic(() => import("react-player").then((mod) => mod.default as ComponentType<any>), {
  ssr: false,
});

export function VideoLesson({ video }: { video: VideoDetail }) {
  const sourceUrl = video.source_url || `https://www.youtube.com/watch?v=${video.provider_id}`;

  return (
    <div className="space-y-4">
      <div className="aspect-video bg-black rounded-lg overflow-hidden">
        <ReactPlayer
          url={sourceUrl}
          playing={false}
          controls={true}
          width="100%"
          height="100%"
        />
      </div>
      <a
        href={sourceUrl}
        target="_blank"
        rel="noopener noreferrer"
        className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
      >
        <ExternalLink className="h-3.5 w-3.5" />
        Watch on YouTube
      </a>
      {video.curator_notes && (
        <div className="border-t pt-4">
          <h3 className="text-sm font-medium mb-2">Curator Notes</h3>
          <TiptapRenderer content={video.curator_notes} />
        </div>
      )}
    </div>
  );
}
