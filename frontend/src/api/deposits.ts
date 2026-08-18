import { z } from "zod";
import { DepositSchema } from "./schemas";
import { client } from "./client";

type Deposit = z.infer<typeof DepositSchema>;

export async function listDeposits(goalId: string): Promise<Deposit> {
  return client<Deposit>(
    `api/goals/${goalId}/deposits`,
    {
      method: "GET",
    },
    DepositSchema,
  );
}

export async function createDeposit(
  goalId: string,
  amount: number,
  note?: string,
): Promise<Deposit> {
  return client(
    `api/goals/${goalId}`,
    {
      method: "POST",
      body: {
        amount,
        note,
      },
    },
    DepositSchema,
  );
}
