"use client";

import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";
import { TiptapRenderer } from "@/components/editor/TiptapRenderer";
import { submitQuizAttempt } from "@/lib/api/lessons";
import type { QuizQuestion, QuizAttemptResult } from "@/lib/types";
import { cn } from "@/lib/utils";
import { CheckCircle, XCircle } from "lucide-react";

export function QuizLesson({
  courseId,
  lessonId,
  questions,
}: {
  courseId: string;
  lessonId: string;
  questions: QuizQuestion[];
}) {
  const [answers, setAnswers] = useState<Record<string, string>>({});
  const [results, setResults] = useState<QuizAttemptResult | null>(null);

  const mutation = useMutation({
    mutationFn: () => submitQuizAttempt(courseId, lessonId, answers),
    onSuccess: (data) => setResults(data),
  });

  const multipleCorrect = (q: QuizQuestion) =>
    (q.options?.filter((o) => o.correct).length ?? 0) > 1;

  const handleSubmit = () => {
    mutation.mutate();
  };

  const handleRetry = () => {
    setAnswers({});
    setResults(null);
  };

  return (
    <div className="space-y-6">
      {questions.map((q, idx) => {
        const result = results?.question_results?.[idx];

        return (
          <div key={q.id} className="border rounded-lg p-4 space-y-3">
            <div className="flex items-start gap-2">
              <span className="text-sm font-medium text-muted-foreground shrink-0">
                {idx + 1}.
              </span>
              <TiptapRenderer content={q.prompt} />
            </div>

            {q.type === "multiple_choice" && q.options && (
              <div className="space-y-2 pl-6">
                {q.options.map((opt) => {
                  const isSelected = answers[q.id] === opt.id;
                  const showCorrect = result && opt.correct;
                  const showIncorrect = result && isSelected && !opt.correct;

                  return (
                    <label
                      key={opt.id}
                      className={cn(
                        "flex items-center gap-2 rounded-md px-3 py-2 text-sm cursor-pointer border transition-colors",
                        isSelected && !result && "border-primary bg-primary/5",
                        showCorrect && "border-green-500 bg-green-50",
                        showIncorrect && "border-red-500 bg-red-50",
                        !isSelected && !result && "border-transparent hover:bg-accent",
                        result && "cursor-default"
                      )}
                    >
                      <input
                        type={multipleCorrect(q) ? "checkbox" : "radio"}
                        name={q.id}
                        value={opt.id}
                        checked={isSelected}
                        onChange={() => !result && setAnswers({ ...answers, [q.id]: opt.id })}
                        disabled={!!result}
                        className="sr-only"
                      />
                      {result && opt.correct && <CheckCircle className="h-4 w-4 text-green-600 shrink-0" />}
                      {result && isSelected && !opt.correct && <XCircle className="h-4 w-4 text-red-600 shrink-0" />}
                      <span>{opt.text}</span>
                    </label>
                  );
                })}
              </div>
            )}

            {q.type === "short_answer" && (
              <div className="pl-6">
                <textarea
                  className="w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm"
                  rows={3}
                  placeholder="Type your answer..."
                  value={answers[q.id] || ""}
                  onChange={(e) => !result && setAnswers({ ...answers, [q.id]: e.target.value })}
                  disabled={!!result}
                />
              </div>
            )}

            {result && q.explanation && (
              <div className="pl-6 border-t pt-2">
                <p className="text-xs font-medium text-muted-foreground mb-1">Explanation</p>
                <TiptapRenderer content={q.explanation} />
              </div>
            )}
          </div>
        );
      })}

      <div className="flex items-center gap-3">
        {!results ? (
          <Button onClick={handleSubmit} disabled={mutation.isPending}>
            {mutation.isPending ? "Submitting..." : "Submit answers"}
          </Button>
        ) : (
          <>
            {results.score !== undefined && (
              <p className="text-sm font-medium">Score: {results.score.toFixed(0)}%</p>
            )}
            <Button onClick={handleRetry} variant="outline">
              Try again
            </Button>
          </>
        )}
      </div>
    </div>
  );
}
