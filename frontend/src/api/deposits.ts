import { DepositSchema, type Deposit } from "./schemas";
import { client } from "./client";

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
