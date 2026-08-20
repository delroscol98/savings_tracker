import { z } from "zod";
import { DepositSchema, type Deposit } from "./schemas";
import { client } from "./client";

export async function listDeposits(goalId: string): Promise<Deposit[]> {
  return client<Deposit[]>(
    `api/goals/${goalId}/deposits`,
    {
      method: "GET",
    },
    z.array(DepositSchema),
  );
}

export async function createDeposit(
  goalId: string,
  amount: number,
  note?: string,
): Promise<Deposit> {
  return client(
    `api/goals/${goalId}/deposits`,
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
