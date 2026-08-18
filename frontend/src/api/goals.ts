import { z } from "zod";
import { GoalSchema, MessageSchema } from "./schemas";
import { client } from "./client";

type Goal = z.infer<typeof GoalSchema>;
type Message = z.infer<typeof MessageSchema>;

export async function listGoals(): Promise<Goal[]> {
  return client<Goal[]>(
    "api/goals",
    {
      method: "GET",
    },
    z.array(GoalSchema),
  );
}

export async function createGoal(
  target: number,
  deadline: Date,
): Promise<Goal> {
  return client<Goal>(
    "api/goals",
    {
      method: "POST",
      body: {
        target,
        deadline: deadline.toISOString(),
      },
    },
    GoalSchema,
  );
}

export async function updateGoal(
  goalId: string,
  target: number,
  deadline: Date,
): Promise<Goal> {
  return client<Goal>(
    `api/goals/${goalId}`,
    {
      method: "PUT",
      body: {
        target,
        deadline: deadline.toISOString(),
      },
    },
    GoalSchema,
  );
}

export async function deleteGoal(goalId: string): Promise<Message> {
  return client<Message>(`api/goals/${goalId}`, {
    method: "DELETE",
  });
}
