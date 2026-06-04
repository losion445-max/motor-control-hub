import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"

export function Layout() {
    return (
        <div className="min-h-screen bg-background p-4 md:p-8">
            <div className="mx-auto max-w-2xl">
                <h1 className="mb-8 text-3xl font-bold">Motor Control HUB SUPER UI v3</h1>

                <Tabs defaultValue="control">
                    <TabsList className="w-full">
                        <TabsTrigger value="control" className="flex-1">Управление</TabsTrigger>
                        <TabsTrigger value="settings" className="flex-1">Настройки</TabsTrigger>
                        <TabsTrigger value="config" className="flex-1">Конфиг</TabsTrigger>
                        <TabsTrigger value="info" className="flex-1">Информация</TabsTrigger>
                    </TabsList>

                    <TabsContent value="settings">Даниил пидарас</TabsContent>
                </Tabs>
            </div>
        </div>
    )
}